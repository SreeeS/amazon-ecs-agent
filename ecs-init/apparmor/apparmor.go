package apparmor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/aaparser"
	aaprofile "github.com/docker/docker/profiles/apparmor"
)

const (
	ECSDefaultProfileName = "ecs-default"
	appArmorProfileDir    = "/etc/apparmor.d"
)

const ecsDefaultProfile = `
#include <tunables/global>

profile ecs-default flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>

  network inet, # Allow IPv4 traffic
  network inet6, # Allow IPv6 traffic

  capability net_admin, # Allow network configuration
  capability sys_admin, # Allow ECS Agent to invoke the setns system call
  
  file,
  umount,
  # Host (privileged) processes may send signals to container processes.
  signal (receive) peer=unconfined,
  # Container processes may send signals amongst themselves.
  signal (send,receive) peer=ecs-default,
  
  # ECS agent requires DBUS send
  dbus (send) bus=system,

  # suppress ptrace denials when using 'docker ps' or using 'ps' inside a container
  ptrace (trace,read,tracedby,readby) peer=ecs-default,
}
`

var (
	isProfileLoaded = aaprofile.IsLoaded
	loadPath        = aaparser.LoadProfile
	createFile      = os.Create
)

// LoadDefaultProfile ensures the default profile to be loaded with the given name.
// Returns nil error if the profile is already loaded.
func LoadDefaultProfile(profileName string) error {
	yes, err := isProfileLoaded(profileName)
	if yes {
		return nil
	}
	if err != nil {
		return err
	}

	f, err := createFile(filepath.Join(appArmorProfileDir, profileName))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(ecsDefaultProfile)
	if err != nil {
		return err
	}
	path := f.Name()

	if err := loadPath(path); err != nil {
		return fmt.Errorf("error loading apparmor profile %s: %w", path, err)
	}
	return nil
}
