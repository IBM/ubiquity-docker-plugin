package block_device_utils

import (
	"os/exec"
	"syscall"
	"github.com/IBM/ubiquity/utils/logs"
)

func (s *impBlockDeviceUtils) CheckFs(mpath string) (bool, error) {
	defer s.logger.Trace(logs.DEBUG)()
	// TODO check first if mpath exist
	needFs := false
	blkidCmd := "blkid"
	if err := s.exec.IsExecutable(blkidCmd); err != nil {
		return false, s.logger.ErrorRet(&commandNotFoundError{blkidCmd, err}, "failed")
	}
	args := []string{blkidCmd, mpath}
	outputBytes, err := s.exec.Execute("sudo", args)
	if err != nil {
		if s.IsExitStatusCode(err, 2) {
			// TODO we can improve it by double check the fs type of this device and maybe log warning if its not the same fstype we expacted
			needFs = true
		} else {
			return false, s.logger.ErrorRet(&commandExecuteError{blkidCmd, err}, "failed")
		}
	}
	s.logger.Info("checked", logs.Args{{"needFs", needFs}, {"mpath", mpath}, {blkidCmd, outputBytes}})
	return needFs, nil
}

func (s *impBlockDeviceUtils) MakeFs(mpath string, fsType string) error {
	defer s.logger.Trace(logs.DEBUG)()
	mkfsCmd := "mkfs"
	if err := s.exec.IsExecutable(mkfsCmd); err != nil {
		return s.logger.ErrorRet(&commandNotFoundError{mkfsCmd, err}, "failed")
	}
	args := []string{mkfsCmd, "-t", fsType, mpath}
	if _, err := s.exec.Execute("sudo", args); err != nil {
		return s.logger.ErrorRet(&commandExecuteError{mkfsCmd, err}, "failed")
	}
	s.logger.Info("created", logs.Args{{"fsType", fsType}, {"mpath", mpath}})
	return nil
}

func (s *impBlockDeviceUtils) IsExitStatusCode(err error, code int) bool {
	defer s.logger.Trace(logs.DEBUG)()
	isExitStatusCode := false
	if status, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := status.ProcessState.Sys().(syscall.WaitStatus); ok {
			isExitStatusCode = waitStatus.ExitStatus() == code
		}
	}
	s.logger.Info("verified", logs.Args{{"isExitStatusCode", isExitStatusCode}, {"code", code}, {"error", err}})
	return isExitStatusCode
}
