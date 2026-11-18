package xkeen

import (
	"errors"
	"fmt"
	"time"
)

var ErrCoreSwitchBusy = errors.New("XKeen уже переключается или перезапускается")

// SwitchCore changes the active XKeen 2.x core. Call from a background job:
// XKeen can take longer than a panel HTTP request to stop or start.
func (d *Detector) SwitchCore(target string) error {
	if target != CoreXray && target != CoreMihomo {
		return errors.New("неизвестное ядро")
	}
	restartMu.Lock()
	if restarting {
		restartMu.Unlock()
		return ErrCoreSwitchBusy
	}
	restarting = true
	restartMu.Unlock()
	defer func() {
		restartMu.Lock()
		restarting = false
		restartMu.Unlock()
		if OnRestartStateChange != nil {
			OnRestartStateChange(false)
		}
	}()
	if OnRestartStateChange != nil {
		OnRestartStateChange(true)
	}
	before := d.freshRuntime()
	if !before.Installed || before.Generation != 2 || (before.Core != CoreXray && before.Core != CoreMihomo) {
		return errors.New("смена ядра поддерживается только в XKeen 2.x")
	}
	if before.Core == target {
		return nil
	}
	if !coreProcessRunning(before.Core) || coreProcessRunning(target) {
		return errors.New("для смены должно работать только текущее ядро XKeen")
	}
	if _, err := TestConfig(before.Dispatcher, target); err != nil {
		Log("[CORE] Config check for %s failed: %v", target, err)
		return fmt.Errorf("конфигурация %s не прошла проверку; текущее ядро не остановлено", target)
	}
	if err := d.activateCore(before, target); err != nil {
		Log("[CORE] Switch %s → %s failed: %v", before.Core, target, err)
		if rollbackErr := d.restoreCore(before, target); rollbackErr != nil {
			Log("[CORE] Rollback to %s failed: %v", before.Core, rollbackErr)
			return fmt.Errorf("не удалось запустить %s или восстановить %s; проверь состояние XKeen", target, before.Core)
		}
		return fmt.Errorf("не удалось переключиться на %s; прежнее ядро %s восстановлено", target, before.Core)
	}
	return nil
}

func (d *Detector) freshRuntime() Runtime {
	d.mu.Lock()
	d.cachedAt = time.Time{}
	d.mu.Unlock()
	d.InvalidateTopology()
	return d.Runtime()
}

func (d *Detector) activateCore(before Runtime, target string) error {
	if _, err := Stop(before.Dispatcher); err != nil {
		return fmt.Errorf("stop: %w", err)
	}
	if coreProcessRunning(before.Core) || coreProcessRunning(target) {
		return errors.New("a proxy process is still running after stop")
	}
	if _, err := runXkeen(before.Dispatcher, quickTimeout, "-"+target); err != nil {
		return fmt.Errorf("select core: %w", err)
	}
	if d.freshRuntime().Core != target {
		return errors.New("XKeen did not select the requested core")
	}
	if _, err := Start(before.Dispatcher); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	return d.verifyCore(before, target, before.Core)
}

func (d *Detector) restoreCore(before Runtime, target string) error {
	current := d.freshRuntime()
	if current.Core != before.Core || coreProcessRunning(target) {
		if _, err := Stop(before.Dispatcher); err != nil && coreProcessRunning(target) {
			return fmt.Errorf("stop failed target: %w", err)
		}
	}
	if coreProcessRunning(target) {
		return errors.New("failed target still running")
	}
	if d.freshRuntime().Core != before.Core {
		if _, err := runXkeen(before.Dispatcher, quickTimeout, "-"+before.Core); err != nil {
			return fmt.Errorf("select previous core: %w", err)
		}
	}
	if d.freshRuntime().Core != before.Core {
		return errors.New("XKeen did not restore the previous selection")
	}
	if !coreProcessRunning(before.Core) {
		if _, err := Start(before.Dispatcher); err != nil {
			return fmt.Errorf("start previous core: %w", err)
		}
	}
	return d.verifyCore(before, before.Core, target)
}

func (d *Detector) verifyCore(before Runtime, wanted, other string) error {
	for i := 0; i < 10; i++ {
		current := d.freshRuntime()
		modeOK := before.Mode != "Hybrid" || current.Mode == "Hybrid"
		if modeOK && current.Core == wanted && coreProcessRunning(wanted) && !coreProcessRunning(other) {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return errors.New("selected core, process, or Hybrid mode did not become ready")
}
