package env

import (
	"os"
	"testing"
)

func createEnv(dir, content string) error {
	if err := os.Chdir(dir); err != nil {
		return err
	}

	f, err := os.Create(dir + "/.env")

	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	if _, err := f.Write([]byte(content)); err != nil {
		return err
	}

	return nil
}

func createEnvByName(name, dir, content string) error {
	if err := os.Chdir(dir); err != nil {
		return err
	}

	f, err := os.Create(name)

	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	if _, err := f.Write([]byte(content)); err != nil {
		return err
	}

	return nil
}

func TestNoFileEnv(t *testing.T) {
	if err := Load(); err == nil {
		t.Fatal(err)
	}
}

func TestLoadEnvEmpty(t *testing.T) {
	dir := t.TempDir()

	if err := createEnv(dir, ""); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)

	if err := Load(); err == nil {
		t.Fatal(err)
	}
}

func TestLoadEnvWithValueEmpty(t *testing.T) {
	dir := t.TempDir()

	content := `
	  KEY=
	`

	if err := createEnv(dir, content); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)

	if err := Load(); err != nil {
		t.Fatal(err)
	}

	value := os.Getenv("KEY")

	if value != "" {
		t.Fatalf("variable KEY = %s must be empty", value)
	}
}

func TestLoadEnvWithValueEmptyWithtrailing(t *testing.T) {
	dir := t.TempDir()

	content := `
	  KEY=""
	`

	if err := createEnv(dir, content); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)

	if err := Load(); err != nil {
		t.Fatal(err)
	}

	value := os.Getenv("KEY")

	if value != "" {
		t.Fatalf("variable KEY = %s must be empty", value)
	}
}

func TestLoadEnvWithKeyEmpty(t *testing.T) {
	dir := t.TempDir()

	content := `
	  =key
	`

	if err := createEnv(dir, content); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)

	if err := Load(); err == nil {
		t.Fatal(err)
	}
}

func TestLoadEnvWithNothingEnvs(t *testing.T) {
	dir := t.TempDir()

	content := ","

	if err := createEnv(dir, content); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)

	if err := Load(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadEnv(t *testing.T) {
	dir := t.TempDir()

	content := `
	  ENVIRONMENT=test
	`

	if err := createEnv(dir, content); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)

	if err := Load(); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("ENVIRONMENT") != "test" {
		t.Fatalf("env ENVIRONMENT must be test but get %s", os.Getenv("ENVIRONMENT"))
	}
}

func TestSetName(t *testing.T) {
	dir := t.TempDir()

	content := `
	  ENVIRONMENT=test
	`

	if err := createEnv(dir, content); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)

	if err := Load(".env"); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("ENVIRONMENT") != "test" {
		t.Fatalf("env ENVIRONMENT must be test but get %s", os.Getenv("ENVIRONMENT"))
	}
}

func TestManyFilesEnv(t *testing.T) {
	files := []struct {
		name, content string
	}{struct {
		name    string
		content string
	}{name: ".env", content: "ENVIRONMENT=test"}, struct {
		name    string
		content string
	}{name: ".env.test", content: "SECRET=testing"},
		struct {
			name    string
			content string
		}{name: ".env.example", content: "SOMETHING=you"}}

	dir := t.TempDir()
	t.Chdir(dir)

	filenames := []string{}

	for _, file := range files {
		if err := createEnvByName(file.name, dir, file.content); err != nil {
			t.Fatal(err)
		}

		filenames = append(filenames, file.name)
	}

	if err := Load(filenames...); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("ENVIRONMENT") != "test" {
		t.Fatal("")
	}

	if os.Getenv("SECRET") != "testing" {
		t.Fatal()
	}

	if os.Getenv("SOMETHING") != "you" {
		t.Fatal("")
	}
}
