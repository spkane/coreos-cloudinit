package system

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestPlaceNetworkUnit(t *testing.T) {
	u := Unit{
		Name:    "50-eth0.network",
		Runtime: true,
		Content: `[Match]
Name=eth47

[Network]
Address=10.209.171.177/19
`,
	}

	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	dst := UnitDestination(&u, dir)
	expectDst := path.Join(dir, "run", "systemd", "network", "50-eth0.network")
	if dst != expectDst {
		t.Fatalf("UnitDestination returned %s, expected %s", dst, expectDst)
	}

	if err := PlaceUnit(&u, dst); err != nil {
		t.Fatalf("PlaceUnit failed: %v", err)
	}

	fi, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if fi.Mode() != os.FileMode(0644) {
		t.Errorf("File has incorrect mode: %v", fi.Mode())
	}

	contents, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	expectContents := `[Match]
Name=eth47

[Network]
Address=10.209.171.177/19
`
	if string(contents) != expectContents {
		t.Fatalf("File has incorrect contents '%s'.\nExpected '%s'", string(contents), expectContents)
	}
}

func TestUnitDestination(t *testing.T) {
	dir := "/some/dir"
	name := "foobar.service"

	u := Unit{
		Name:   name,
		DropIn: false,
	}

	dst := UnitDestination(&u, dir)
	expectDst := path.Join(dir, "etc", "systemd", "system", "foobar.service")
	if dst != expectDst {
		t.Errorf("UnitDestination returned %s, expected %s", dst, expectDst)
	}

	u.DropIn = true

	dst = UnitDestination(&u, dir)
	expectDst = path.Join(dir, "etc", "systemd", "system", "foobar.service.d", cloudConfigDropIn)
	if dst != expectDst {
		t.Errorf("UnitDestination returned %s, expected %s", dst, expectDst)
	}
}

func TestPlaceMountUnit(t *testing.T) {
	u := Unit{
		Name:    "media-state.mount",
		Runtime: false,
		Content: `[Mount]
What=/dev/sdb1
Where=/media/state
`,
	}

	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	dst := UnitDestination(&u, dir)
	expectDst := path.Join(dir, "etc", "systemd", "system", "media-state.mount")
	if dst != expectDst {
		t.Fatalf("UnitDestination returned %s, expected %s", dst, expectDst)
	}

	if err := PlaceUnit(&u, dst); err != nil {
		t.Fatalf("PlaceUnit failed: %v", err)
	}

	fi, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if fi.Mode() != os.FileMode(0644) {
		t.Errorf("File has incorrect mode: %v", fi.Mode())
	}

	contents, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	expectContents := `[Mount]
What=/dev/sdb1
Where=/media/state
`
	if string(contents) != expectContents {
		t.Fatalf("File has incorrect contents '%s'.\nExpected '%s'", string(contents), expectContents)
	}
}

func TestMachineID(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	os.Mkdir(path.Join(dir, "etc"), os.FileMode(0755))
	ioutil.WriteFile(path.Join(dir, "etc", "machine-id"), []byte("node007\n"), os.FileMode(0444))

	if MachineID(dir) != "node007" {
		t.Fatalf("File has incorrect contents")
	}
}

func TestMaskUnit(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Ensure mask works with units that do not currently exist
	if err := MaskUnit("foo.service", dir); err != nil {
		t.Fatalf("Unable to mask new unit: %v", err)
	}
	fooPath := path.Join(dir, "etc", "systemd", "system", "foo.service")
	fooTgt, err := os.Readlink(fooPath)
	if err != nil {
		t.Fatalf("Unable to read link", err)
	}
	if fooTgt != "/dev/null" {
		t.Fatalf("unit not masked, got unit target", fooTgt)
	}

	// Ensure mask works with unit files that already exist
	barPath := path.Join(dir, "etc", "systemd", "system", "bar.service")
	if _, err := os.Create(barPath); err != nil {
		t.Fatalf("Error creating new unit file: %v", err)
	}
	if err := MaskUnit("bar.service", dir); err != nil {
		t.Fatalf("Unable to mask existing unit: %v", err)
	}
	barTgt, err := os.Readlink(barPath)
	if err != nil {
		t.Fatalf("Unable to read link", err)
	}
	if barTgt != "/dev/null" {
		t.Fatalf("unit not masked, got unit target", barTgt)
	}
}
