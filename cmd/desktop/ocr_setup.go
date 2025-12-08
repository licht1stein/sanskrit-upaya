package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/licht1stein/sanskrit-upaya/pkg/gcloud"
	"github.com/licht1stein/sanskrit-upaya/pkg/ocr"
)

// SetupState tracks the current state of the setup wizard
type SetupState int

const (
	StateCheckingGcloud SetupState = iota
	StateGcloudNotFound
	StateAuthenticating
	StateCreatingProject
	StateEnablingAPI
	StateSettingADC
	StateSettingQuota
	StateEnableBilling
	StateVerifying
	StateComplete
	StateError
)

// OCRSetupWizard manages the OCR setup UI
type OCRSetupWizard struct {
	window       fyne.Window
	parentWindow fyne.Window
	onComplete   func() // Called when setup succeeds

	// UI elements
	stepLabel     *widget.Label
	progressLabel *widget.Label
	outputText    *widget.Entry
	rerunBtn      *widget.Button
	cancelBtn     *widget.Button
	continueBtn   *widget.Button

	// State
	state     SetupState
	projectID string
	mu        sync.Mutex
	cancelled bool
}

// NewOCRSetupWizard creates a new setup wizard
func NewOCRSetupWizard(app fyne.App, parentWindow fyne.Window, onComplete func()) *OCRSetupWizard {
	w := &OCRSetupWizard{
		parentWindow: parentWindow,
		onComplete:   onComplete,
	}

	w.window = app.NewWindow("OCR Setup Wizard")
	w.window.Resize(fyne.NewSize(600, 450))
	w.window.SetFixedSize(true)

	w.buildUI()
	return w
}

func (w *OCRSetupWizard) buildUI() {
	// Step indicator
	w.stepLabel = widget.NewLabel("Checking Google Cloud CLI...")
	w.stepLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Progress description
	w.progressLabel = widget.NewLabel("")
	w.progressLabel.Wrapping = fyne.TextWrapWord

	// Console output area
	w.outputText = widget.NewMultiLineEntry()
	w.outputText.Wrapping = fyne.TextWrapWord
	w.outputText.Disable() // Read-only
	w.outputText.SetMinRowsVisible(12)

	// Buttons
	w.rerunBtn = widget.NewButton("Re-run Setup", func() {
		w.startSetup()
	})
	w.rerunBtn.Hide()

	w.continueBtn = widget.NewButton("Continue", func() {
		w.continueAfterBilling()
	})
	w.continueBtn.Hide()

	w.cancelBtn = widget.NewButton("Cancel", func() {
		w.cancel()
	})

	buttonRow := container.NewHBox(w.rerunBtn, w.continueBtn, w.cancelBtn)

	// Layout
	content := container.NewBorder(
		container.NewVBox(
			w.stepLabel,
			w.progressLabel,
			widget.NewSeparator(),
		),
		container.NewVBox(
			widget.NewSeparator(),
			container.NewCenter(buttonRow),
		),
		nil, nil,
		container.NewScroll(w.outputText),
	)

	w.window.SetContent(container.NewPadded(content))
}

// Show displays the wizard and starts the setup process
func (w *OCRSetupWizard) Show() {
	w.window.Show()
	w.startSetup()
}

func (w *OCRSetupWizard) cancel() {
	w.mu.Lock()
	w.cancelled = true
	w.mu.Unlock()
	w.window.Close()
}

func (w *OCRSetupWizard) appendOutput(line string) {
	fyne.Do(func() {
		current := w.outputText.Text
		if current != "" {
			current += "\n"
		}
		w.outputText.SetText(current + line)
		// Scroll to bottom by moving cursor
		w.outputText.CursorRow = strings.Count(w.outputText.Text, "\n")
	})
}

func (w *OCRSetupWizard) clearOutput() {
	fyne.Do(func() {
		w.outputText.SetText("")
	})
}

func (w *OCRSetupWizard) setStep(step string) {
	fyne.Do(func() {
		w.stepLabel.SetText(step)
	})
}

func (w *OCRSetupWizard) setProgress(text string) {
	fyne.Do(func() {
		w.progressLabel.SetText(text)
	})
}

func (w *OCRSetupWizard) showRerunButton() {
	fyne.Do(func() {
		w.rerunBtn.Show()
		w.continueBtn.Hide()
	})
}

func (w *OCRSetupWizard) showContinueButton() {
	fyne.Do(func() {
		w.continueBtn.Show()
		w.rerunBtn.Hide()
	})
}

func (w *OCRSetupWizard) hideButtons() {
	fyne.Do(func() {
		w.rerunBtn.Hide()
		w.continueBtn.Hide()
	})
}

func (w *OCRSetupWizard) isCancelled() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.cancelled
}

func (w *OCRSetupWizard) startSetup() {
	w.mu.Lock()
	w.cancelled = false
	w.mu.Unlock()

	w.clearOutput()
	w.hideButtons()

	go w.runSetupSteps()
}

func (w *OCRSetupWizard) runSetupSteps() {
	// Step 1: Check gcloud CLI
	w.setStep("Step 1 of 6: Checking Google Cloud CLI...")
	w.setProgress("Looking for gcloud command...")

	if !gcloud.IsInstalled() {
		w.handleGcloudNotFound()
		return
	}
	w.appendOutput("✓ Google Cloud CLI found")

	if w.isCancelled() {
		return
	}

	// Step 2: Authenticate gcloud
	w.setStep("Step 2 of 6: Authenticating...")

	if !gcloud.IsAuthenticated() {
		w.setProgress("A browser window will open. Please log in with your Google account.")
		w.appendOutput("$ gcloud auth login")
		w.appendOutput("Opening browser for authentication...")

		result := <-gcloud.RunCommandAsync(w.appendOutput, "auth", "login")
		if !result {
			w.handleError("Authentication failed", "Please try again. If browser didn't open, check the output above for the URL.")
			return
		}
	}
	w.appendOutput("✓ Authenticated with Google Cloud")

	if w.isCancelled() {
		return
	}

	// Step 3: Check if OCR already works
	w.setStep("Step 3 of 6: Checking existing credentials...")
	w.setProgress("Verifying if Vision API is already accessible...")

	ctx := context.Background()
	if gcloud.HasApplicationDefaultCredentials() {
		if err := ocr.CheckCredentials(ctx); err == nil {
			w.appendOutput("✓ Vision API already configured and accessible!")
			w.handleComplete()
			return
		}
	}
	w.appendOutput("Setting up Vision API access...")

	// Get project ID
	var err error
	w.projectID, err = gcloud.GetOrCreateOCRProjectID()
	if err != nil {
		w.handleError("Failed to get project ID", err.Error())
		return
	}

	if w.isCancelled() {
		return
	}

	// Step 4: Create project if needed
	w.setStep("Step 4 of 6: Setting up GCP project...")
	w.setProgress(fmt.Sprintf("Checking project '%s'...", w.projectID))

	if !gcloud.ProjectExists(w.projectID) {
		w.appendOutput(fmt.Sprintf("$ gcloud projects create %s", w.projectID))
		w.appendOutput("Creating project (this may take a moment)...")

		result := <-gcloud.RunCommandAsync(w.appendOutput, "projects", "create", w.projectID, "--name=Sanskrit Upaya OCR")
		if !result {
			w.handleProjectCreationError()
			return
		}
		w.appendOutput("✓ Project created")
	} else {
		w.appendOutput(fmt.Sprintf("✓ Project '%s' already exists", w.projectID))
	}

	if w.isCancelled() {
		return
	}

	// Step 5: Enable Vision API
	w.setStep("Step 5 of 6: Enabling Vision API...")
	w.setProgress("Enabling the Cloud Vision API for your project...")
	w.appendOutput("$ gcloud services enable vision.googleapis.com")

	result := <-gcloud.RunCommandAsync(w.appendOutput, "services", "enable", "vision.googleapis.com", "--project="+w.projectID)
	if !result {
		w.handleAPIEnableError()
		return
	}
	w.appendOutput("✓ Vision API enabled")

	if w.isCancelled() {
		return
	}

	// Step 6: Set up ADC
	w.setStep("Step 6 of 6: Setting up credentials...")
	w.setProgress("A browser window will open. Please log in again to set up Application Default Credentials.")
	w.appendOutput("$ gcloud auth application-default login")

	result = <-gcloud.RunCommandAsync(w.appendOutput, "auth", "application-default", "login")
	if !result {
		w.handleError("ADC setup failed", "Failed to set up Application Default Credentials. Please try again.")
		return
	}
	w.appendOutput("✓ Application Default Credentials configured")

	if w.isCancelled() {
		return
	}

	// Set quota project
	w.appendOutput("Setting quota project...")
	w.appendOutput(fmt.Sprintf("$ gcloud auth application-default set-quota-project %s", w.projectID))

	result = <-gcloud.RunCommandAsync(w.appendOutput, "auth", "application-default", "set-quota-project", w.projectID)
	if !result {
		w.handleError("Quota project setup failed", "Failed to set quota project. Please try again.")
		return
	}
	w.appendOutput("✓ Quota project configured")

	if w.isCancelled() {
		return
	}

	// Verify
	w.setStep("Verifying setup...")
	w.setProgress("Testing Vision API access...")

	if err := ocr.CheckCredentials(ctx); err != nil {
		if errors.Is(err, ocr.ErrBillingDisabled) {
			w.handleBillingRequired()
			return
		}
		w.handleError("Verification failed", err.Error())
		return
	}

	w.handleComplete()
}

func (w *OCRSetupWizard) handleGcloudNotFound() {
	w.setStep("Google Cloud CLI Not Found")
	w.setProgress("The Google Cloud CLI (gcloud) is required for OCR setup.")

	w.appendOutput("❌ gcloud command not found")
	w.appendOutput("")
	w.appendOutput("Please install it from:")
	w.appendOutput(gcloud.GetInstallURL())
	w.appendOutput("")
	w.appendOutput("After installing:")
	w.appendOutput("1. Restart your terminal/application")
	w.appendOutput("2. Click 'Re-run Setup' below")

	w.showRerunButton()

	// Open browser to install page
	if err := gcloud.OpenBrowser(gcloud.GetInstallURL()); err == nil {
		w.appendOutput("")
		w.appendOutput("(Opening installation page in browser...)")
	}
}

func (w *OCRSetupWizard) handleProjectCreationError() {
	w.setStep("Project Creation Failed")
	w.setProgress("Could not create the GCP project. This usually means you need to accept Google Cloud terms.")

	w.appendOutput("")
	w.appendOutput("❌ Project creation failed")
	w.appendOutput("")
	w.appendOutput("You may need to:")
	w.appendOutput("1. Accept Google Cloud terms of service")
	w.appendOutput("2. Have a billing account (Vision API has a free tier)")
	w.appendOutput("")
	w.appendOutput("Opening Google Cloud Console...")

	if err := gcloud.OpenBrowser(gcloud.GetConsoleURL()); err == nil {
		w.appendOutput("Please accept any terms, then click 'Re-run Setup'")
	}

	w.showRerunButton()
}

func (w *OCRSetupWizard) handleAPIEnableError() {
	w.setStep("Vision API Enable Failed")
	w.setProgress("Could not enable the Vision API. You may need to enable billing first.")

	w.appendOutput("")
	w.appendOutput("❌ Could not enable Vision API")
	w.appendOutput("")
	w.appendOutput("This usually requires billing to be enabled.")
	w.appendOutput("(Don't worry - there's a free tier of 1000 images/month)")
	w.appendOutput("")

	billingURL := gcloud.GetBillingURL(w.projectID)
	if err := gcloud.OpenBrowser(billingURL); err == nil {
		w.appendOutput("Opening billing setup page...")
	}
	w.appendOutput("")
	w.appendOutput("After enabling billing, click 'Re-run Setup'")

	w.showRerunButton()
}

func (w *OCRSetupWizard) handleBillingRequired() {
	w.setStep("Billing Required")
	w.setProgress("Almost done! You need to enable billing for the project.")

	w.appendOutput("")
	w.appendOutput("⚠️ Billing not enabled for project")
	w.appendOutput("")
	w.appendOutput("The Vision API requires billing to be enabled.")
	w.appendOutput("Don't worry - there's a FREE tier:")
	w.appendOutput("• 1000 images/month FREE")
	w.appendOutput("• Then $1.50 per 1000 images")
	w.appendOutput("")

	billingURL := gcloud.GetBillingURL(w.projectID)
	if err := gcloud.OpenBrowser(billingURL); err == nil {
		w.appendOutput("Opening billing setup page...")
	} else {
		w.appendOutput(fmt.Sprintf("Please open: %s", billingURL))
	}
	w.appendOutput("")
	w.appendOutput("After enabling billing, click 'Continue' below.")

	w.showContinueButton()
}

func (w *OCRSetupWizard) continueAfterBilling() {
	w.hideButtons()
	w.setStep("Verifying billing...")
	w.setProgress("Checking if billing is now enabled...")

	go func() {
		ctx := context.Background()
		if err := ocr.CheckCredentials(ctx); err != nil {
			w.appendOutput("")
			w.appendOutput(fmt.Sprintf("❌ Still not working: %v", err))
			w.appendOutput("")
			w.appendOutput("If you just enabled billing, wait a minute and try again.")
			w.showContinueButton()
			return
		}
		w.handleComplete()
	}()
}

func (w *OCRSetupWizard) handleError(title, message string) {
	w.setStep(title)
	w.setProgress(message)
	w.appendOutput("")
	w.appendOutput(fmt.Sprintf("❌ %s", title))
	w.appendOutput(message)
	w.showRerunButton()
}

func (w *OCRSetupWizard) handleComplete() {
	w.setStep("Setup Complete!")
	w.setProgress("OCR is ready to use. Free tier: 1000 images/month.")

	w.appendOutput("")
	w.appendOutput("✓ Vision API accessible")
	w.appendOutput("")
	w.appendOutput("=== Setup Complete! ===")
	w.appendOutput("")
	w.appendOutput("You can now use OCR to recognize Sanskrit text from images.")
	w.appendOutput("Free tier: 1000 images/month, then $1.50/1000")

	fyne.Do(func() {
		w.cancelBtn.SetText("Close")
		w.cancelBtn.OnTapped = func() {
			w.window.Close()
			if w.onComplete != nil {
				w.onComplete()
			}
		}
	})
}

// ShowOCRSetupDialog shows a simple confirmation dialog before starting setup
func ShowOCRSetupDialog(parent fyne.Window, app fyne.App, onComplete func()) {
	content := widget.NewLabel(
		"OCR requires Google Cloud Vision API credentials.\n\n" +
			"The setup wizard will:\n" +
			"1. Check for Google Cloud CLI\n" +
			"2. Authenticate with your Google account\n" +
			"3. Create a GCP project for OCR\n" +
			"4. Enable the Vision API\n" +
			"5. Set up credentials\n\n" +
			"Free tier: 1000 images/month",
	)
	content.Wrapping = fyne.TextWrapWord

	dlg := dialog.NewCustomConfirm("OCR Setup Required", "Start Setup", "Cancel", content, func(ok bool) {
		if ok {
			wizard := NewOCRSetupWizard(app, parent, onComplete)
			wizard.Show()
		}
	}, parent)
	dlg.Resize(fyne.NewSize(400, 300))
	dlg.Show()
}
