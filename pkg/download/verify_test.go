package download

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/licht1stein/sanskrit-upaya/pkg/search"
)

// withExpected substitutes the checksum/size Download and CheckDatabase
// verify against, restoring the real values when the test ends.
func withExpected(t *testing.T, checksum string, size int64, legacy ...string) {
	t.Helper()
	oldSum, oldSize, oldLegacy := expectedChecksum, expectedSize, LegacyChecksums
	expectedChecksum, expectedSize, LegacyChecksums = checksum, size, legacy
	t.Cleanup(func() { expectedChecksum, expectedSize, LegacyChecksums = oldSum, oldSize, oldLegacy })
}

func sha(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func fileSHA(t *testing.T, path string) string {
	t.Helper()
	s, err := computeFileChecksum(path)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// buildRollbackDB creates a small dictionary DB in rollback-journal mode,
// like the published file, and returns its path.
func buildRollbackDB(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "sanskrit.db")
	db, err := search.OpenForBulkInsert(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.InitSchemaForBulkInsert(); err != nil {
		t.Fatal(err)
	}
	bi, err := db.NewBulkInserter()
	if err != nil {
		t.Fatal(err)
	}
	if err := bi.InsertDict("mw", "Monier-Williams", "sa", "en", true); err != nil {
		t.Fatal(err)
	}
	id, err := bi.InsertArticle("mw", "धर्म law, duty")
	if err != nil {
		t.Fatal(err)
	}
	if err := bi.InsertWord("धर्म", "धर्म", id, "mw"); err != nil {
		t.Fatal(err)
	}
	if err := bi.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := db.RebuildFTS(); err != nil {
		t.Fatal(err)
	}
	db.Close()
	return path
}

func TestCheckDatabase_StaysValidAfterOpen(t *testing.T) {
	dir := t.TempDir()
	path := buildRollbackDB(t, dir)
	before := fileSHA(t, path)
	withExpected(t, before, 0)

	if got := checkDatabaseAt(path); got != DatabaseValid {
		t.Fatalf("fresh file: got %v, want DatabaseValid", got)
	}

	// Opening the DB the way the app does modifies it in place (WAL mode).
	db, err := search.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	res, err := db.Search("धर्म", search.ModeExact, nil)
	if err != nil || len(res) == 0 {
		t.Fatalf("search after open: %d results, err %v", len(res), err)
	}
	db.Close()

	if after := fileSHA(t, path); after == before {
		t.Log("search.Open did not modify the file; marker not exercised")
	}

	// Next startup must not ask for a re-download.
	if got := checkDatabaseAt(path); got != DatabaseValid {
		t.Fatalf("after open: got %v, want DatabaseValid (would re-download every launch)", got)
	}
}

func TestCheckDatabase_ModifiedWithoutMarkerNeedsUpdate(t *testing.T) {
	dir := t.TempDir()
	path := buildRollbackDB(t, dir)
	withExpected(t, fileSHA(t, path), 0)

	db, err := search.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	if fileSHA(t, path) == expectedChecksum {
		t.Skip("search.Open did not modify the file")
	}
	if got := checkDatabaseAt(path); got != DatabaseNeedsUpdate {
		t.Fatalf("got %v, want DatabaseNeedsUpdate", got)
	}
}

func TestCheckDatabase_LegacyChecksumAccepted(t *testing.T) {
	dir := t.TempDir()
	path := buildRollbackDB(t, dir)
	withExpected(t, "0000000000000000000000000000000000000000000000000000000000000000", 0, fileSHA(t, path))

	if got := checkDatabaseAt(path); got != DatabaseValid {
		t.Fatalf("got %v, want DatabaseValid", got)
	}
	if m := readVerifiedMarker(path); m != LegacyChecksums[0] {
		t.Fatalf("marker = %q, want legacy checksum", m)
	}
}

func TestCheckDatabase_StaleMarkerRehashes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sanskrit.db")
	os.WriteFile(path, []byte(sqliteMagic+"old database"), 0644)
	writeVerifiedMarker(path, "previous-release-checksum")
	withExpected(t, sha([]byte("new database")), 0)

	if got := checkDatabaseAt(path); got != DatabaseNeedsUpdate {
		t.Fatalf("got %v, want DatabaseNeedsUpdate", got)
	}
}

func TestCheckDatabase_MarkerButNotSQLite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sanskrit.db")
	content := []byte("not a database")
	os.WriteFile(path, content, 0644)
	withExpected(t, sha([]byte("something else")), 0)
	writeVerifiedMarker(path, expectedChecksum)

	if got := checkDatabaseAt(path); got != DatabaseNeedsUpdate {
		t.Fatalf("got %v, want DatabaseNeedsUpdate", got)
	}
}

func serve(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(HeaderName) != AppSecret {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestDownload_VerifiesAndWritesMarker(t *testing.T) {
	body := []byte(sqliteMagic + "published database")
	withExpected(t, sha(body), int64(len(body)))
	srv := serve(t, body)
	dir := t.TempDir()
	path := filepath.Join(dir, "sanskrit.db")
	writeVerifiedMarker(path, "stale")
	os.WriteFile(path+"-wal", []byte("stale wal"), 0644)

	if err := downloadTo(srv.URL, path, nil); err != nil {
		t.Fatalf("downloadTo: %v", err)
	}
	if got := fileSHA(t, path); got != expectedChecksum {
		t.Fatalf("file checksum %s", got)
	}
	if m := readVerifiedMarker(path); m != expectedChecksum {
		t.Fatalf("marker = %q", m)
	}
	if _, err := os.Stat(path + "-wal"); !os.IsNotExist(err) {
		t.Fatal("stale -wal file not removed")
	}
	if got := checkDatabaseAt(path); got != DatabaseValid {
		t.Fatalf("CheckDatabase after download = %v", got)
	}
}

func TestDownload_ChecksumMismatch(t *testing.T) {
	body := []byte(sqliteMagic + "unexpected database")
	withExpected(t, sha([]byte("something else")), int64(len(body)))
	srv := serve(t, body)
	path := filepath.Join(t.TempDir(), "sanskrit.db")

	if err := downloadTo(srv.URL, path, nil); err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("database file should not exist after failed download")
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("temp file should be cleaned up")
	}
}

func TestDownload_SizeMismatch(t *testing.T) {
	body := []byte(sqliteMagic + "published database")
	withExpected(t, sha(body), int64(len(body))+1)
	srv := serve(t, body)
	path := filepath.Join(t.TempDir(), "sanskrit.db")

	if err := downloadTo(srv.URL, path, nil); err == nil {
		t.Fatal("expected size mismatch error")
	}
}

func TestPinnedValues(t *testing.T) {
	if len(ExpectedChecksum) != 64 {
		t.Fatalf("ExpectedChecksum must be a hex SHA256, got %q", ExpectedChecksum)
	}
	if ExpectedSize <= 0 {
		t.Fatal("ExpectedSize must be set")
	}
	for _, c := range LegacyChecksums {
		if c == ExpectedChecksum {
			t.Fatal("ExpectedChecksum should not also be listed as legacy")
		}
	}
}
