const BASE = 'http://localhost:3000/api/v1';

export class ApiClient {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  static async login(login: string, password: string): Promise<{ token: string; user: Record<string, unknown> }> {
    const res = await fetch(`${BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ login, password }),
    });
    if (!res.ok) throw new Error(`Login failed: ${res.status}`);
    return res.json();
  }

  private async request(method: string, path: string, body?: unknown): Promise<unknown> {
    const opts: RequestInit = {
      method,
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.token}`,
      },
    };
    if (body) opts.body = JSON.stringify(body);
    const res = await fetch(`${BASE}${path}`, opts);
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`API ${method} ${path} failed (${res.status}): ${text}`);
    }
    if (res.status === 204) return {};
    return res.json();
  }

  async get(path: string) { return this.request('GET', path); }
  async getRaw(path: string): Promise<Response> {
    return fetch(`${BASE}${path}`, {
      headers: { Authorization: `Bearer ${this.token}` },
    });
  }
  async post(path: string, body: unknown) { return this.request('POST', path, body); }
  async put(path: string, body: unknown) { return this.request('PUT', path, body); }
  async del(path: string) { return this.request('DELETE', path); }

  // Entity helpers
  async createAccount(overrides: Record<string, unknown> = {}) {
    return this.post('/accounts', { name: `E2E Account ${Date.now()}`, access: 'Public', ...overrides }) as Promise<Record<string, unknown>>;
  }

  async createContact(overrides: Record<string, unknown> = {}) {
    return this.post('/contacts', { first_name: 'E2E', last_name: `Contact ${Date.now()}`, access: 'Public', ...overrides }) as Promise<Record<string, unknown>>;
  }

  async createLead(overrides: Record<string, unknown> = {}) {
    return this.post('/leads', { first_name: 'E2E', last_name: `Lead ${Date.now()}`, access: 'Public', ...overrides }) as Promise<Record<string, unknown>>;
  }

  async createOpportunity(overrides: Record<string, unknown> = {}) {
    return this.post('/opportunities', { name: `E2E Opp ${Date.now()}`, stage: 'prospecting', access: 'Public', ...overrides }) as Promise<Record<string, unknown>>;
  }

  async createCampaign(overrides: Record<string, unknown> = {}) {
    return this.post('/campaigns', { name: `E2E Campaign ${Date.now()}`, access: 'Public', ...overrides }) as Promise<Record<string, unknown>>;
  }

  async createTask(overrides: Record<string, unknown> = {}) {
    return this.post('/tasks', { name: `E2E Task ${Date.now()}`, bucket: 'due_asap', ...overrides }) as Promise<Record<string, unknown>>;
  }

  async deleteEntity(type: string, id: number) {
    try { await this.del(`/${type}/${id}`); } catch { /* ignore cleanup errors */ }
  }

  /** Ensure a non-admin user exists for authorization tests. */
  async ensureNonAdminUser(): Promise<{ username: string; password: string }> {
    const username = 'e2e_nonadmin';
    const password = 'Dem0P@ssword!!';
    try {
      await this.post('/admin/users', {
        username,
        email: 'e2e_nonadmin@fatfreecrm.local',
        password,
        password_confirmation: password,
        admin: false,
      });
    } catch {
      // User may already exist — that's fine
    }
    return { username, password };
  }
}
