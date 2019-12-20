package rscsrv

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	colorSuccess   = color.New(color.FgGreen)
	colorError     = color.New(color.FgRed)
	colorWarning   = color.New(color.FgYellow)
	colorHighlight = color.New(color.FgWhite, color.Bold)
	colorMuted     = color.New(color.FgHiBlack)

	colorBold = color.New(color.Bold)
)

// Success formats text as
var (
	formatSuccess   = colorSuccess.SprintfFunc()
	formatError     = colorError.SprintfFunc()
	formatWarning   = colorWarning.SprintfFunc()
	formatHighlight = colorHighlight.SprintfFunc()
	formatMuted     = colorMuted.SprintfFunc()

	formatBold = colorSuccess.SprintfFunc()
)

type ColorStarterReporter struct{}

var DefaultColorStarterReporter = &ColorStarterReporter{}

const colorTitleL1 string = "    %-27s\n"

func (reporter *ColorStarterReporter) printL1f(format string, args ...interface{}) {
	fmt.Printf(colorTitleL1, fmt.Sprintf(format, args...))
}

func (reporter *ColorStarterReporter) BeforeBegin(service Service) {
	fmt.Printf("%s\n", formatHighlight(service.Name()))
}

func (reporter *ColorStarterReporter) BeforeLoadConfiguration(service Configurable) {
	reporter.printL1f("Loading configuration ...")
}

func (reporter *ColorStarterReporter) printError(err error) {
	var t string
	if err != nil {
		t = formatBold(formatError("Error"))
		reporter.printL1f("> [%s]: %s", t, err)
		return
	}
	t = formatBold(formatSuccess("OK"))
	reporter.printL1f("> [%s]", t)
}

func (reporter *ColorStarterReporter) AfterLoadConfiguration(service Configurable, conf interface{}, err error) {
	reporter.printError(err)
}

func (reporter *ColorStarterReporter) BeforeApplyConfiguration(service Configurable) {
	reporter.printL1f("Applying configuration ...")
}

func (reporter *ColorStarterReporter) AfterApplyConfiguration(service Configurable, conf interface{}, err error) {
	reporter.printError(err)
}

func (reporter *ColorStarterReporter) BeforeStart(service Service) {
	reporter.printL1f("Starting ...")
}

func (reporter *ColorStarterReporter) AfterStart(service Service, err error) {
	reporter.printError(err)
}

func (reporter *ColorStarterReporter) BeforeStop(service Service) {
	reporter.printL1f("Stopping ...")
}

func (reporter *ColorStarterReporter) AfterStop(service Service, err error) {
	reporter.printError(err)
}

// ReportRetrier is called whenever a service is started or not. If the
// service is successfully started, err will be nil, otherwise not.
func (reporter *ColorStarterReporter) ReportRetrier(retrier *StartRetrier, err error) error {
	if err != nil {
		reporter.printL1f("Retrier > [%s]: Try %d: %s", formatError("Error"), retrier.Try+1, err)
	}
	return err
}
