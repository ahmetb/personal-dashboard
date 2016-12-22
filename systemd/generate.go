// Generates unit/timer files for systemd and prints the commands
// to execute.
package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type task struct {
	name    string
	fixture string
	cmd     string
}

var fixture = []task{
	{"reporting", "*:0/20", "reporting"},
	{"twitter-followers", "*:59:00", "twitter followers"},
	{"twitter-tweets", "*:59:00", "twitter tweets"},
	{"github", "*:59:00", "github"},
	{"jawbone-sleeps", "*:0/15", "jawbone sleeps 3"},
	{"jawbone-steps", "*:0/15", "jawbone steps 3"},
	{"jawbone-caffeine", "*:0/15", "jawbone caffeine 3"},
	{"lastfm", "*:0/20", "lastfm"},
	{"wakatime", "*:0/20", "wakatime"},
}

func main() {
	for _, svc := range fixture {
		serviceFn := fmt.Sprintf("pd-%s.service", svc.name)
		timerFn := fmt.Sprintf("pd-%s.timer", svc.name)

		serviceB := []byte(fmt.Sprintf(`[Unit]
Description=Scheduled personal dashboard task for '%s'

[Service]
Type=oneshot
ExecStart=/usr/bin/docker run --rm -v /home/core/personal-dashboard/config.toml:/etc/personal-dashboard/config.toml -v /home/core/personal-dashboard/google-credentials.json:/etc/personal-dashboard/google-credentials.json  -e GOOGLE_APPLICATION_CREDENTIALS=/etc/personal-dashboard/google-credentials.json ahmet/personal-dashboard %s
`, svc.name, svc.cmd))

		if err := ioutil.WriteFile(serviceFn, serviceB, 0644); err != nil {
			panic(err)
		}

		timerB := []byte(fmt.Sprintf(`[Unit]
Description=Timer for %s

[Timer]
OnCalendar=%s
AccuracySec=5s
Persistent=true

[Install]
WantedBy=timers.target
`, serviceFn, svc.fixture))

		if err := ioutil.WriteFile(timerFn, timerB, 0644); err != nil {
			panic(err)
		}
	}
	cmd := []string{
		"scp ./*.{timer,service} SERVER:/etc/systemd/system",
		"systemctl enable /etc/systemd/system/pd-*.timer",
		"systemctl start /etc/systemd/system/pd-*.timer",
		"systemctl list-timers --all | grep pd-"}

	fmt.Println(strings.Join(cmd, "\n"))
}
