import { auth } from './auth.svelte.js';

const API_BASE = '/api';

// --- HELPER FUNCTIONS ---

async function request(endpoint, options = {}) {
  const isFormData = options.body instanceof FormData;

  const headers = {
    ...(!isFormData && { 'Content-Type': 'application/json' }),
    ...(auth.token ? { Authorization: `Bearer ${auth.token}` } : {}),
    ...options.headers,
  };

  const config = { ...options, headers };

  const res = await fetch(`${API_BASE}${endpoint}`, config);

  if (res.status === 204) return { success: true };

  const contentType = res.headers.get('content-type') || '';
  let data = null;
  if (contentType.includes('application/json')) {
    data = await res.json();
  } else {
    const text = await res.text();
    data = text ? { message: text } : null;
  }

  // For login endpoint, let caller handle 401 so we can show proper message
  if (res.status === 401) {
    if (!endpoint.startsWith('/auth/login')) {
      auth.logout();
    }
    throw new Error(data?.error || data?.message || 'Unauthorized');
  }

  if (!res.ok) throw new Error(data?.error || data?.message || `Request failed (${res.status})`);
  return data;
}

async function download(endpoint, params = {}) {
  const query = new URLSearchParams(params).toString();
  const res = await fetch(`${API_BASE}${endpoint}${query ? `?${query}` : ''}`, {
    headers: auth.token ? { Authorization: `Bearer ${auth.token}` } : {},
  });
  if (!res.ok) throw new Error('Download failed');
  const blob = await res.blob();
  const disposition = res.headers.get('Content-Disposition') || '';
  const match = disposition.match(/filename="?([^";]+)"?/);
  const filename = match?.[1] || 'download.xlsx';
  return { blob, filename };
}

// --- API OBJECT EXPORT ---

export const api = {
  // ==========================================
  // GENERIC HELPERS (Supaya api.post jalan)
  // ==========================================
  async get(endpoint, params = {}) {
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) searchParams.append(key, String(value));
    });
    const query = searchParams.toString();
    return request(`${endpoint}${query ? `?${query}` : ''}`);
  },

  async post(endpoint, body) {
    return request(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
    });
  },

  async put(endpoint, body) {
    return request(endpoint, {
      method: 'PUT',
      body: JSON.stringify(body),
    });
  },

  async delete(endpoint) {
    return request(endpoint, { method: 'DELETE' });
  },

  async patch(endpoint, body) {
    return request(endpoint, {
      method: 'PATCH',
      body: JSON.stringify(body),
    });
  },

  // ==========================================
  // AUTH
  // ==========================================
  async login(email, password, totpCode = null) {
    // Go backend biasanya expect snake_case "totp_code"
    const data = await request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password, totp_code: totpCode }),
    });
    if (data?.data?.token) {
      auth.login(data.data.user, data.data.token);
    }
    return data;
  },
  async getGoogleOAuthUrl(redirectPath = '/') {
    const res = await request(`/auth/google?redirect_path=${encodeURIComponent(redirectPath)}`);
    return res.data?.url || null;
  },
  async requestPasswordReset(email) {
    return request('/auth/password/forgot', { method: 'POST', body: JSON.stringify({ email }) });
  },
  async resetPassword(payload) {
    return request('/auth/password/reset', { method: 'POST', body: JSON.stringify(payload) });
  },
  async getCurrentUser() {
    return request('/auth/me');
  },
  async setup2FA() {
    return request('/auth/2fa/setup', { method: 'POST' });
  },
  async verify2FA(code) {
    return request('/auth/2fa/verify', { method: 'POST', body: JSON.stringify({ code }) });
  },
  async disable2FA() {
    return request('/auth/2fa/disable', { method: 'POST' });
  },
  async logout() {
    try {
      await request('/auth/logout', { method: 'POST' });
    } finally {
      auth.logout();
    }
  },

  // ==========================================
  // ADMIN & INTERN MANAGEMENT
  // ==========================================
  
  // Dashboard Statistic
  async getAdminDashboard() {
    return request('/dashboard/admin'); 
  },
  
  async updateInternStatus(id, status) {
    return request(`/interns/${id}`, { 
        method: 'PUT', 
        body: JSON.stringify({ status }) 
    });
  },

  // ==========================================
  // INTERNS (CRUD sesuai intern.go)
  // ==========================================
  async getInterns(params = {}) {
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) searchParams.append(key, String(value));
    });
    const query = searchParams.toString();
    return request(`/interns${query ? `?${query}` : ''}`);
  },
  async getIntern(id) {
    return request(`/interns/${id}`);
  },
  async createIntern(payload) {
    return request('/interns', { method: 'POST', body: JSON.stringify(payload) });
  },
  async updateIntern(id, payload) {
    return request(`/interns/${id}`, { method: 'PUT', body: JSON.stringify(payload) });
  },
  async deleteIntern(id) {
    return request(`/interns/${id}`, { method: 'DELETE' });
  },

  // ==========================================
  // PROFILE
  // ==========================================
  async getProfile() {
    return request('/profile');
  },
  async updateProfile(formData) {
    return request('/profile', {
      method: 'PUT',
      body: formData, // FormData handle content-type automatically in helper
    });
  },
  async updatePassword(payload) {
    return request('/profile/password', { method: 'PUT', body: JSON.stringify(payload) });
  },

  // ==========================================
  // SUPERVISORS
  // ==========================================
  async getSupervisors(params = {}) {
    const query = new URLSearchParams(params).toString();
    return request(`/admin/supervisors${query ? `?${query}` : ''}`);
  },
  async createSupervisor(payload) {
    return request('/admin/supervisors', { method: 'POST', body: JSON.stringify(payload) });
  },
  async updateSupervisor(id, payload) {
    return request(`/admin/supervisors/${id}`, { method: 'PUT', body: JSON.stringify(payload) });
  },
  async deleteSupervisor(id) {
    return request(`/admin/supervisors/${id}`, { method: 'DELETE' });
  },
  async approveSupervisor(id) {
    return request(`/admin/supervisors/${id}/approve`, { method: 'POST' });
  },
  async rejectSupervisor(id) {
    return request(`/admin/supervisors/${id}/reject`, { method: 'POST' });
  },

  // ==========================================
  // TASKS
  // ==========================================
  async getTasks(arg1, arg2) {
    let params = {};
    if (arg1 && typeof arg1 === 'object') {
        params = { ...arg1 };
    } else {
        if (arg1) params.intern_id = arg1; 
        params.page = arg2 || 1;
    }
    
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
        if (value !== null && value !== undefined) {
            searchParams.append(key, String(value));
        }
    });
    
    const query = searchParams.toString();

    if (params.intern_id) {
        return request(`/tasks/intern/${params.intern_id}?page=${params.page || 1}`);
    }
    
    return request(`/tasks${query ? `?${query}` : ''}`);
  },

  async getTask(id) {
    return request(`/tasks/${id}`);
  },
  async createTask(payload) {
    return request('/tasks', { method: 'POST', body: JSON.stringify(payload) });
  },
  async updateTask(id, payload) {
    return request(`/tasks/${id}`, { method: 'PUT', body: JSON.stringify(payload) });
  },
  async deleteTask(id) {
    return request(`/tasks/${id}`, { method: 'DELETE' });
  },
  async submitTask(id, payload) {
    return request(`/tasks/${id}/submit`, { method: 'POST', body: JSON.stringify(payload) });
  },
  async reviewTask(id, payload) {
    return request(`/tasks/${id}/review`, { method: 'POST', body: JSON.stringify(payload) });
  },
  async updateTaskStatus(id, payload) {
    return request(`/tasks/${id}/status`, { method: 'POST', body: JSON.stringify(payload) });
  },
  async uploadTaskAttachment(id, file) {
    const formData = new FormData();
    formData.append('file', file);
    // Use manual fetch for attachment to avoid request helper override content-type
    const res = await fetch(`${API_BASE}/tasks/${id}/attachments`, {
      method: 'POST',
      headers: auth.token ? { Authorization: `Bearer ${auth.token}` } : {},
      body: formData,
    });
    return res.json();
  },
  async searchInterns(q) {
    return request(`/tasks/search-interns?q=${encodeURIComponent(q)}`);
  },
  async getTaskAssignments(params = {}) {
    const query = new URLSearchParams(params).toString();
    return request(`/task-assignments${query ? `?${query}` : ''}`);
  },
  async getTaskAssignment(id) {
    return request(`/task-assignments/${id}`);
  },

  // ==========================================
  // ATTENDANCE
  // ==========================================
  async checkIn(latitude, longitude, reason = null) {
    const body = { latitude, longitude };
    if (reason) body.reason = reason;
    return request('/attendance/checkin', { method: 'POST', body: JSON.stringify(body) });
  },
  async checkOut(latitude, longitude) {
    return request('/attendance/checkout', { method: 'POST', body: JSON.stringify({ latitude, longitude }) });
  },
  async submitPermission(payload) {
    const formData = new FormData();
    Object.entries(payload).forEach(([key, value]) => {
      if (value !== null && value !== undefined) formData.append(key, value);
    });
    // Use manual fetch for formdata
    const res = await fetch(`${API_BASE}/attendance/permission`, {
      method: 'POST',
      headers: auth.token ? { Authorization: `Bearer ${auth.token}` } : {},
      body: formData,
    });
    return res.json();
  },
  async getTodayAttendance() {
    return request('/attendance/today');
  },
  async getAttendance(params = {}) {
    const query = new URLSearchParams(params).toString();
    return request(`/attendance${query ? `?${query}` : ''}`);
  },
  async getAttendanceById(id) {
    return request(`/attendance/${id}`);
  },
  async getAttendanceByIntern(id, params = {}) {
    const query = new URLSearchParams(params).toString();
    return request(`/attendance/intern/${id}${query ? `?${query}` : ''}`);
  },

  // ==========================================
  // LEAVES
  // ==========================================
  async getLeaveRequests(params = {}) {
    const query = new URLSearchParams(params).toString();
    return request(`/leaves${query ? `?${query}` : ''}`);
  },
  async createLeaveRequest(payload) {
    return request('/leaves', { method: 'POST', body: JSON.stringify(payload) });
  },
  async approveLeave(id) {
    return request(`/leaves/${id}/approve`, { method: 'POST' });
  },
  async rejectLeave(id) {
    return request(`/leaves/${id}/reject`, { method: 'POST' });
  },

  // ==========================================
  // ASSESSMENTS
  // ==========================================
  async getAssessments(params = {}) {
    const query = new URLSearchParams(params).toString();
    return request(`/assessments${query ? `?${query}` : ''}`);
  },
  async createAssessment(payload) {
    return request('/assessments', { method: 'POST', body: JSON.stringify(payload) });
  },
  async updateAssessment(id, payload) {
    return request(`/assessments/${id}`, { method: 'PUT', body: JSON.stringify(payload) });
  },
  async deleteAssessment(id) {
    return request(`/assessments/${id}`, { method: 'DELETE' });
  },

  // ==========================================
  // REPORTS
  // ==========================================
  async getReports(params = {}) {
    const query = new URLSearchParams(params).toString();
    return request(`/reports${query ? `?${query}` : ''}`);
  },
  async createReport(payload) {
    return request('/reports', { method: 'POST', body: JSON.stringify(payload) });
  },
  async addReportFeedback(id, feedback) {
    return request(`/reports/${id}/feedback`, { method: 'POST', body: JSON.stringify({ feedback }) });
  },
  async getInternReport(id) {
    return request(`/reports/intern/${id}`);
  },

  // ==========================================
  // NOTIFICATIONS
  // ==========================================
  async getNotifications(params = {}) {
    const query = new URLSearchParams(params).toString();
    return request(`/notifications${query ? `?${query}` : ''}`);
  },
  async markNotificationRead(id) {
    return request(`/notifications/${id}/read`, { method: 'POST' });
  },
  async markAllNotificationsRead() {
    return request('/notifications/mark-all-read', { method: 'POST' });
  },
  async deleteNotification(id) {
    return request(`/notifications/${id}`, { method: 'DELETE' });
  },

  // ==========================================
  // EXPORT/IMPORT
  // ==========================================
  async exportInterns(params = {}) {
    return download('/export/interns', params);
  },
  async exportAttendances(params = {}) {
    return download('/export/attendances', params);
  },
  async exportTasks(params = {}) {
    return download('/export/tasks', params);
  },
  async downloadImportTemplate() {
    return download('/import/template');
  },
  async importInterns(file, supervisorId = null) {
    const formData = new FormData();
    formData.append('file', file);
    if (supervisorId) formData.append('supervisor_id', supervisorId);
    const res = await fetch(`${API_BASE}/import/interns`, {
      method: 'POST',
      headers: auth.token ? { Authorization: `Bearer ${auth.token}` } : {},
      body: formData,
    });
    return res.json();
  },

  // ==========================================
  // ANALYTICS / DASHBOARD
  // ==========================================
  async getInternDashboardData() {
    return request('/analytics/dashboard/intern');
  },
  async getWeeklyAttendanceStats() {
    return request('/analytics/attendance/weekly');
  },
  async getTaskBreakdown() {
    return request('/analytics/tasks/breakdown');
  },
  async getAttendancePercentage() {
    return request('/analytics/attendance/percentage');
  },

  // Settings
  async getSettings() {
    return request('/settings');
  },
  async updateSettings(payload) {
    return request('/settings', { method: 'POST', body: JSON.stringify(payload) });
  },

  // Analytics Specific
  async getWeeklyTrends(internId, weekOffset = 0) {
    return request(`/analytics/trends/weekly/${internId}?week_offset=${weekOffset}`);
  },
  async getCheckInPatterns(internId, days = 30) {
    return request(`/analytics/patterns/checkin/${internId}?days=${days}`);
  },
  async getPerformanceInsights(internId) {
    return request(`/analytics/insights/${internId}`);
  },

  // Holidays
  async getHolidays(params = {}) {
    const query = new URLSearchParams(params).toString();
    return request(`/holidays${query ? `?${query}` : ''}`);
  },
};

// Export request if needed elsewhere
export { request };
