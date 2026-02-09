const API_BASE = '/api';

class ApiClient {
  constructor() {
    this.token = localStorage.getItem('token');
  }

  setToken(token) {
    this.token = token;
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
  }

  getHeaders() {
    const headers = {
      'Content-Type': 'application/json',
    };
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }
    return headers;
  }

  async request(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;
    const config = {
      ...options,
      headers: {
        ...this.getHeaders(),
        ...options.headers,
      },
    };

    try {
      const response = await fetch(url, config);
      
      // Handle non-JSON responses
      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        return { success: true };
      }

      const data = await response.json();

      if (!response.ok) {
        if (response.status === 401) {
          this.setToken(null);
        }
        throw new Error(data.error || data.message || 'Request failed');
      }

      return data;
    } catch (error) {
      console.error('API Error:', error);
      throw error;
    }
  }

  // ==========================================
  // Auth endpoints
  // ==========================================
  async login(email, password, totpCode = null) {
    const data = await this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password, totp_code: totpCode }),
    });
    if (data.data && data.data.token) {
      this.setToken(data.data.token);
    }
    return data;
  }

  async register(userData) {
    return this.request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
  }

  async getCurrentUser() {
    return this.request('/auth/me');
  }

  async logout() {
    try {
      await this.request('/auth/logout', { method: 'POST' });
    } catch (error) {
      // Ignore logout errors
    } finally {
      this.setToken(null);
    }
  }

  // ==========================================
  // ADMIN & INTERN MANAGEMENT (DITAMBAHKAN)
  // ==========================================
  
  async getAdminDashboard() {
    return this.request('/admin/dashboard');
  }

  async getInterns(params = {}) {
    // Convert params object to query string
    const queryString = new URLSearchParams(params).toString();
    // Gunakan endpoint admin jika ada, atau fallback ke /interns
    const endpoint = `/admin/interns${queryString ? '?' + queryString : ''}`;
    return this.request(endpoint);
  }

  async createIntern(data) {
    return this.request('/admin/interns', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  // FUNGSI PENTING UNTUK TOMBOL APPROVE
  async updateInternStatus(id, status) {
    return this.request(`/admin/interns/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status })
    });
  }

  // ==========================================
  // Attendance endpoints
  // ==========================================
  async checkIn(latitude, longitude, reason = null) {
    const body = { latitude, longitude };
    if (reason) {
      body.reason = reason;
    }
    return this.request('/attendance/checkin', {
      method: 'POST',
      body: JSON.stringify(body),
    });
  }

  async checkOut(latitude, longitude) {
    return this.request('/attendance/checkout', {
      method: 'POST',
      body: JSON.stringify({ latitude, longitude }),
    });
  }

  async getTodayAttendance() {
    return this.request('/attendance/today');
  }

  async getAttendanceHistory(internId, page = 1, limit = 30) {
    return this.request(`/attendance/intern/${internId}?page=${page}&limit=${limit}`);
  }

  // ==========================================
  // Analytics endpoints
  // ==========================================
  async getWeeklyTrends(internId, weekOffset = 0) {
    return this.request(`/analytics/trends/weekly/${internId}?week_offset=${weekOffset}`);
  }

  async getCheckInPatterns(internId, days = 30) {
    return this.request(`/analytics/patterns/checkin/${internId}?days=${days}`);
  }

  async getPerformanceInsights(internId) {
    return this.request(`/analytics/insights/${internId}`);
  }

  // ==========================================
  // Task endpoints
  // ==========================================
  // Updated to support object params {status: 'submitted'}
  async getTasks(arg1 = {}, arg2 = 1) {
    let endpoint = '/tasks';
    
    // Cek apakah arg1 adalah params object atau internId string/null
    if (arg1 && typeof arg1 === 'object') {
        const queryString = new URLSearchParams(arg1).toString();
        endpoint = `/tasks${queryString ? '?' + queryString : ''}`;
    } else {
        // Logic Lama (internId, page)
        const internId = arg1;
        const page = arg2;
        endpoint = internId 
          ? `/tasks/intern/${internId}?page=${page}` 
          : `/tasks?page=${page}`;
    }
    return this.request(endpoint);
  }

  async getTask(taskId) {
    return this.request(`/tasks/${taskId}`);
  }

  async createTask(taskData) {
    return this.request('/tasks', {
      method: 'POST',
      body: JSON.stringify(taskData),
    });
  }

  async uploadTaskAttachment(taskId, file) {
    const formData = new FormData();
    formData.append('file', file);

    return fetch(`${API_BASE}/tasks/${taskId}/attachments`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
      },
      body: formData,
    }).then(res => res.json());
  }

  // ==========================================
  // Leave endpoints
  // ==========================================
  async getLeaveRequests(internId = null) {
    const endpoint = internId 
      ? `/leaves/intern/${internId}` 
      : `/leaves`;
    return this.request(endpoint);
  }

  async createLeaveRequest(leaveData) {
    return this.request('/leaves', {
      method: 'POST',
      body: JSON.stringify(leaveData),
    });
  }

  async approveLeave(leaveId) {
    return this.request(`/leaves/${leaveId}/approve`, {
      method: 'POST',
      body: JSON.stringify({ status: 'approved' }),
    });
  }

  async rejectLeave(leaveId) {
    return this.request(`/leaves/${leaveId}/reject`, {
      method: 'POST',
      body: JSON.stringify({ status: 'rejected' }),
    });
  }

  // ==========================================
  // Assessment endpoints
  // ==========================================
  async getAssessments(internId = null) {
    const endpoint = internId 
      ? `/assessments/intern/${internId}` 
      : `/assessments`;
    return this.request(endpoint);
  }

  async createAssessment(assessmentData) {
    return this.request('/assessments', {
      method: 'POST',
      body: JSON.stringify(assessmentData),
    });
  }
}

export const api = new ApiClient();