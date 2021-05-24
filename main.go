package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/getlantern/systray"
	"glnt.co.kr/launcher/command"
	"glnt.co.kr/launcher/icon"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	Start = "실행중"
	Stop  = "종료됨"
)

var (
	bar *widget.ProgressBarInfinite

	serviceNames = []string{"ocr", "gpms", "relay"}
	isInit       = true
)

func main() {
	// 한글 폰트 설정.
	os.Setenv("FYNE_FONT", "NanumBarunGothic.ttf")

	// 서비스 실행 체크.
	if runErr := RunCheck(); runErr != nil {
		fmt.Println(runErr)
		os.Exit(1)
	}

	go verifyAllServices()

	// 새로운 앱 실행
	a := app.New()

	// 윈도우 생성
	window := CreateWindow(a)
	window.Show()
	window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("파일",
			fyne.NewMenuItem("Quit", func() {
				fyne.CurrentApp().Quit()
			}),
		),
		fyne.NewMenu("설정 프로그램",
			fyne.NewMenuItem("DABIT", func() { command.ServiceStart("dabit") }),
			fyne.NewMenuItem("WIZNET", func() { command.ServiceStart("wiznet") }),
		),
	))

	// tray 실행
	trayRun(window)

	// 앱 실행
	a.Run()
}

func RunCheck() error {
	output, ce, _ := command.Pipeline(command.TaskList(), command.FindStr("glnt"))
	if len(ce) > 0 {
		return errors.New(string(ce))
	}

	serviceCount := strings.Count(string(output), "\n")
	if serviceCount > 1 {
		return errors.New("<!> program is running")
	}

	return nil
}

func verifyAllServices() {
	for _, v := range serviceNames {
		err := command.ServiceCheck(v)
		if err != nil {
			continue
		}

		command.ServiceStart(v)

		if v == "gpms" {
			time.Sleep(5 * time.Second)
		}
	}
	bar.Hide()
	isInit = false
}

// CreateWindow 윈도우 생성.
func CreateWindow(app fyne.App) fyne.Window {
	window := app.NewWindow("GL&T Launcher")

	bar = widget.NewProgressBarInfinite()

	window.SetIcon(&fyne.StaticResource{StaticContent: icon.MonitoringPng})
	window.SetContent(
		container.NewBorder(
			container.NewVBox(
				createHeaderView(),
				//widget.NewSeparator(),
			),
			bar, nil, nil,
			container.NewVBox(
				//listWidget,
				createLow("ocr"),
				createLow("gpms"),
				createLow("relay"),
			),
		),
	)

	window.Resize(fyne.Size{Height: 230, Width: 360})
	window.CenterOnScreen()
	window.SetFixedSize(true)
	window.SetCloseIntercept(func() {
		window.Hide()
	})

	return window
}

// CreateHeaderView 상단 이름, 상태, 시작/종료 표시
func createHeaderView() *fyne.Container {
	return container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("이름", 0, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("상태", 0, fyne.TextStyle{Bold: true}),
		container.NewHBox(widget.NewLabelWithStyle("시작 / 종료", 1, fyne.TextStyle{Bold: true})),
	)
}

func createLow(name string) *fyne.Container {
	statusLabel := widget.NewLabel(Stop)
	go func(sName string, label *widget.Label) {
		t := time.NewTicker(time.Second)
		for range t.C {
			taskCheck(sName, label)
		}
	}(name, statusLabel)

	startBtn := widget.NewButton("시작", func() {
		if statusLabel.Text == Start {
			return
		}
		if !bar.Running() {
			bar.Show()
			command.ServiceStart(name)
		}
	})

	stopBtn := widget.NewButton("종료", func() {
		if statusLabel.Text == Stop {
			return
		}
		if !bar.Running() {
			bar.Show()
			command.ServiceStop(name)
		}
	})

	return container.NewGridWithColumns(3,
		widget.NewLabel(strings.ToUpper(name)),
		statusLabel,
		container.NewHBox(
			startBtn,
			stopBtn,
		),
	)
}

func taskCheck(name string, label *widget.Label) {
	if isInit {
		return
	}

	err := command.ServiceCheck(name)

	if err != nil {
		if label.Text != Start {
			label.SetText(Start)
			bar.Hide()
		}
	} else {
		if label.Text != Stop {
			label.SetText(Stop)
			bar.Hide()
		}
	}
}

func trayRun(window fyne.Window) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		systray.Run(func() {
			systray.SetIcon(icon.MonitoringIco)
			systray.SetTitle("GLNT Launcher")
			systray.SetTooltip("GLNT Launcher")

			show := systray.AddMenuItem("보이기", "화면 보이기")
			systray.AddSeparator()
			exit := systray.AddMenuItem("종료", "종료")

			wg.Done()

			go func() {
				for {
					select {
					case <-show.ClickedCh:
						window.Show()
					case <-exit.ClickedCh:
						window.Close()
					}
				}
			}()
		}, func() {
			fmt.Println("EXIT")
		})
	}()
	wg.Wait()
}
