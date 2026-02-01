// go_spin UI - Alpine.js Application
function app() {
    return {
        // State
        activeTab: 'containers',
        containers: [],
        groups: [],
        schedules: [],
        error: '',
        success: '',
        
        // Modals
        showContainerModal: false,
        showGroupModal: false,
        showScheduleModal: false,
        
        // Editing flags
        editingContainer: false,
        editingGroup: false,
        editingSchedule: false,
        
        // Forms
        containerForm: {
            name: '',
            friendly_name: '',
            url: '',
            running: false,
            active: true
        },
        groupForm: {
            name: '',
            container: [],
            active: true
        },
        scheduleForm: {
            id: '',
            target: '',
            targetType: 'container',
            timers: []
        },
        
        // Day names for display
        dayNames: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'],
        
        // API base URL (same origin)
        apiBase: '',
        
        // Initialize
        async init() {
            await this.loadAll();
        },
        
        async loadAll() {
            await Promise.all([
                this.loadContainers(),
                this.loadGroups(),
                this.loadSchedules()
            ]);
        },
        
        // ==================== CONTAINERS ====================
        async loadContainers() {
            try {
                const res = await fetch(`${this.apiBase}/containers`);
                if (!res.ok) throw new Error(await res.text());
                this.containers = await res.json();
            } catch (e) {
                this.showError('Failed to load containers: ' + e.message);
            }
        },
        
        openContainerModal(container = null) {
            if (container) {
                this.editingContainer = true;
                this.containerForm = {
                    name: container.name,
                    friendly_name: container.friendly_name,
                    url: container.url,
                    running: container.running || false,
                    active: container.active || false
                };
            } else {
                this.editingContainer = false;
                this.containerForm = {
                    name: '',
                    friendly_name: '',
                    url: '',
                    running: false,
                    active: true
                };
            }
            this.showContainerModal = true;
        },
        
        async saveContainer() {
            try {
                const payload = {
                    name: this.containerForm.name,
                    friendly_name: this.containerForm.friendly_name,
                    url: this.containerForm.url,
                    running: this.containerForm.running,
                    active: this.containerForm.active
                };
                
                const res = await fetch(`${this.apiBase}/container`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || 'Save failed');
                }
                
                this.containers = await res.json();
                this.showContainerModal = false;
                this.showSuccess('Container saved successfully');
            } catch (e) {
                this.showError('Failed to save container: ' + e.message);
            }
        },
        
        async deleteContainer(name) {
            if (!confirm(`Delete container "${name}"?`)) return;
            try {
                const res = await fetch(`${this.apiBase}/container/${encodeURIComponent(name)}`, {
                    method: 'DELETE'
                });
                if (!res.ok) throw new Error(await res.text());
                this.containers = await res.json();
                // Refresh schedules tab because schedules targeting this container were removed server-side
                await this.loadSchedules();
                this.showSuccess('Container deleted');
            } catch (e) {
                this.showError('Failed to delete container: ' + e.message);
            }
        },
        
        async startContainer(name) {
            try {
                const res = await fetch(`${this.apiBase}/runtime/${encodeURIComponent(name)}/start`, {
                    method: 'POST'
                });
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || 'Start failed');
                }
                this.showSuccess(`Container "${name}" started`);
                await this.loadContainers();
            } catch (e) {
                this.showError('Failed to start container: ' + e.message);
            }
        },
        
        async stopContainer(name) {
            try {
                const res = await fetch(`${this.apiBase}/runtime/${encodeURIComponent(name)}/stop`, {
                    method: 'POST'
                });
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || 'Stop failed');
                }
                this.showSuccess(`Container "${name}" stopped`);
                await this.loadContainers();
            } catch (e) {
                this.showError('Failed to stop container: ' + e.message);
            }
        },
        
        // ==================== GROUPS ====================
        async loadGroups() {
            try {
                const res = await fetch(`${this.apiBase}/groups`);
                if (!res.ok) throw new Error(await res.text());
                this.groups = await res.json();
            } catch (e) {
                this.showError('Failed to load groups: ' + e.message);
            }
        },
        
        openGroupModal(group = null) {
            if (group) {
                this.editingGroup = true;
                this.groupForm = {
                    name: group.name,
                    container: [...(group.container || [])],
                    active: group.active || false
                };
            } else {
                this.editingGroup = false;
                this.groupForm = {
                    name: '',
                    container: [],
                    active: true
                };
            }
            this.showGroupModal = true;
        },
        
        toggleGroupContainer(containerName) {
            const idx = this.groupForm.container.indexOf(containerName);
            if (idx === -1) {
                this.groupForm.container.push(containerName);
            } else {
                this.groupForm.container.splice(idx, 1);
            }
        },
        
        async saveGroup() {
            try {
                const payload = {
                    name: this.groupForm.name,
                    container: this.groupForm.container,
                    active: this.groupForm.active
                };
                
                const res = await fetch(`${this.apiBase}/group`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || 'Save failed');
                }
                
                this.groups = await res.json();
                this.showGroupModal = false;
                this.showSuccess('Group saved successfully');
            } catch (e) {
                this.showError('Failed to save group: ' + e.message);
            }
        },
        
        async deleteGroup(name) {
            if (!confirm(`Delete group "${name}"?`)) return;
            try {
                const res = await fetch(`${this.apiBase}/group/${encodeURIComponent(name)}`, {
                    method: 'DELETE'
                });
                if (!res.ok) throw new Error(await res.text());
                this.groups = await res.json();
                // Refresh schedules tab because schedules targeting this group were removed server-side
                await this.loadSchedules();
                this.showSuccess('Group deleted');
            } catch (e) {
                this.showError('Failed to delete group: ' + e.message);
            }
        },
        
        // ==================== SCHEDULES ====================
        async loadSchedules() {
            try {
                const res = await fetch(`${this.apiBase}/schedules`);
                if (!res.ok) throw new Error(await res.text());
                this.schedules = await res.json();
            } catch (e) {
                this.showError('Failed to load schedules: ' + e.message);
            }
        },
        
        openScheduleModal(schedule = null) {
            if (schedule) {
                this.editingSchedule = true;
                this.scheduleForm = {
                    id: schedule.id,
                    target: schedule.target,
                    targetType: schedule.targetType,
                    timers: (schedule.timers || []).map(t => ({
                        startTime: t.startTime,
                        stopTime: t.stopTime,
                        days: [...(t.days || [])],
                        active: t.active || false
                    }))
                };
            } else {
                this.editingSchedule = false;
                this.scheduleForm = {
                    id: this.generateId(),
                    target: '',
                    targetType: 'container',
                    timers: []
                };
            }
            this.showScheduleModal = true;
        },
        
        generateId() {
            return `${Date.now()}-${Math.floor(Math.random() * 10000)}`;
        },
        
        addTimer() {
            this.scheduleForm.timers.push({
                startTime: '08:00',
                stopTime: '18:00',
                days: [1, 2, 3, 4, 5], // Mon-Fri default
                active: true
            });
        },
        
        removeTimer(idx) {
            this.scheduleForm.timers.splice(idx, 1);
        },
        
        toggleTimerDay(timerIdx, dayIdx) {
            const timer = this.scheduleForm.timers[timerIdx];
            const pos = timer.days.indexOf(dayIdx);
            if (pos === -1) {
                timer.days.push(dayIdx);
                timer.days.sort((a, b) => a - b);
            } else {
                timer.days.splice(pos, 1);
            }
        },
        
        formatDays(days) {
            if (!days || days.length === 0) return 'No days';
            if (days.length === 7) return 'Every day';
            const weekdays = [1, 2, 3, 4, 5];
            const weekend = [0, 6];
            if (weekdays.every(d => days.includes(d)) && days.length === 5) return 'Weekdays';
            if (weekend.every(d => days.includes(d)) && days.length === 2) return 'Weekend';
            return days.map(d => this.dayNames[d]).join(', ');
        },
        
        async saveSchedule() {
            try {
                // Build timers with required active field
                const timers = this.scheduleForm.timers.map(t => ({
                    startTime: t.startTime,
                    stopTime: t.stopTime,
                    days: t.days,
                    active: t.active
                }));
                
                const payload = {
                    id: this.scheduleForm.id,
                    target: this.scheduleForm.target,
                    targetType: this.scheduleForm.targetType,
                    timers: timers
                };
                
                const res = await fetch(`${this.apiBase}/schedule`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || 'Save failed');
                }
                
                this.schedules = await res.json();
                this.showScheduleModal = false;
                this.showSuccess('Schedule saved successfully');
            } catch (e) {
                this.showError('Failed to save schedule: ' + e.message);
            }
        },
        
        async deleteSchedule(id) {
            if (!confirm(`Delete schedule "${id}"?`)) return;
            try {
                const res = await fetch(`${this.apiBase}/schedule/${encodeURIComponent(id)}`, {
                    method: 'DELETE'
                });
                if (!res.ok) throw new Error(await res.text());
                this.schedules = await res.json();
                this.showSuccess('Schedule deleted');
            } catch (e) {
                this.showError('Failed to delete schedule: ' + e.message);
            }
        },
        
        // ==================== UTILITIES ====================
        showError(msg) {
            this.error = msg;
            this.success = '';
            setTimeout(() => this.error = '', 5000);
        },
        
        showSuccess(msg) {
            this.success = msg;
            this.error = '';
            setTimeout(() => this.success = '', 3000);
        }
    };
}
