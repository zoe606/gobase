// Command runtimecheck verifies filesystem access inside the runtime image.
package main

import (
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-boilerplate/pkg/project"
)

func main() {
	profile := flag.String("profile", string(project.Full), "runtime profile: full or minimal")
	flag.Parse()
	if err := check(project.Profile(*profile)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Non-root identity, included runtime files, certificates, and writable paths verified")
}

func check(profile project.Profile) error {
	if err := profile.Validate(); err != nil {
		return err
	}
	if err := checkIdentity(); err != nil {
		return err
	}
	patterns := []string{"/config/*.yaml"}
	directories := []string{os.TempDir()}
	if profile == project.Full {
		patterns = append(patterns, "/migrations/*.sql")
		directories = append(directories, "/uploads")
	}
	for _, pattern := range patterns {
		if err := checkReadable(pattern); err != nil {
			return err
		}
	}
	if err := checkCertificates(); err != nil {
		return err
	}
	for _, dir := range directories {
		if err := checkWritable(dir); err != nil {
			return err
		}
	}
	return nil
}

func checkIdentity() error {
	if os.Geteuid() != 65532 || os.Getegid() != 65532 {
		return fmt.Errorf("unexpected runtime identity: %d:%d", os.Geteuid(), os.Getegid())
	}
	status, err := os.ReadFile("/proc/1/status")
	if err != nil {
		return err
	}
	for _, field := range []string{"Uid", "Gid"} {
		if !strings.Contains(string(status), field+":\t65532\t65532\t65532\t65532") {
			return fmt.Errorf("PID 1 does not run with %s 65532", field)
		}
	}
	return nil
}

func checkReadable(pattern string) error {
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("no files matching %s", pattern)
	}
	for _, path := range paths {
		if _, err := os.ReadFile(path); err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
	}
	return nil
}

func checkCertificates() error {
	bundle, err := os.ReadFile("/etc/ssl/certs/ca-certificates.crt")
	if err != nil {
		return err
	}
	if !x509.NewCertPool().AppendCertsFromPEM(bundle) {
		return fmt.Errorf("system certificate bundle contains no certificates")
	}
	_, err = x509.SystemCertPool()
	return err
}

func checkWritable(dir string) error {
	f, err := os.CreateTemp(dir, "runtime-check-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err := f.WriteString("runtime-check"); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
