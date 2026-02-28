export const INITIAL_COLLECTIONS = [
  {
    id: 1, name: 'Auth Service', tags: ['REST', 'OAuth'],
    updated_at: '2026-02-21', default_method: 'POST', accent_color: '#7c3aed', pattern: 'waves',
    requests: [
      { id: 101, user_id: 1, collection_id: 1, name: 'Login',         method: 'POST', url: 'https://api.example.com/auth/login',   headers: [{ k: 'Content-Type', v: 'application/json' }], params: [], body: '{\n  "email": "",\n  "password": ""\n}', auth: 'No Auth', token: '', created_at: '2026-02-21', updated_at: '2026-02-21' },
      { id: 102, user_id: 1, collection_id: 1, name: 'Refresh Token', method: 'POST', url: 'https://api.example.com/auth/refresh', headers: [], params: [], body: '{\n  "refreshToken": ""\n}',                                                   auth: 'Bearer',  token: '', created_at: '2026-02-21', updated_at: '2026-02-21' },
      { id: 103, user_id: 1, collection_id: 1, name: 'Get Profile',   method: 'GET',  url: 'https://api.example.com/auth/me',      headers: [], params: [], body: '',                                                                            auth: 'Bearer',  token: '', created_at: '2026-02-21', updated_at: '2026-02-21' },
      { id: 104, user_id: 1, collection_id: 1, name: 'Logout',        method: 'POST', url: 'https://api.example.com/auth/logout',  headers: [], params: [], body: '',                                                                            auth: 'Bearer',  token: '', created_at: '2026-02-21', updated_at: '2026-02-21' },
    ],
  },
  {
    id: 2, name: 'Payment Gateway', tags: ['REST', 'Stripe'],
    updated_at: '2026-02-20', default_method: 'POST', accent_color: '#0ea5e9', pattern: 'grid',
    requests: [
      { id: 201, user_id: 1, collection_id: 2, name: 'Create Payment Intent', method: 'POST', url: 'https://api.stripe.com/v1/payment_intents', headers: [{ k: 'Authorization', v: 'Bearer sk_test_...' }], params: [], body: '{\n  "amount": 1000,\n  "currency": "usd"\n}', auth: 'Bearer', token: '', created_at: '2026-02-20', updated_at: '2026-02-20' },
      { id: 202, user_id: 1, collection_id: 2, name: 'List Charges',          method: 'GET',  url: 'https://api.stripe.com/v1/charges',          headers: [], params: [{ k: 'limit', v: '10' }], body: '', auth: 'Bearer', token: '', created_at: '2026-02-20', updated_at: '2026-02-20' },
      { id: 203, user_id: 1, collection_id: 2, name: 'Refund',                method: 'POST', url: 'https://api.stripe.com/v1/refunds',          headers: [], params: [], body: '{\n  "charge": ""\n}', auth: 'Bearer', token: '', created_at: '2026-02-20', updated_at: '2026-02-20' },
    ],
  },
  {
    id: 3, name: 'User Profiles API', tags: ['GraphQL'],
    updated_at: '2026-02-18', default_method: 'GET', accent_color: '#10b981', pattern: 'dots',
    requests: [
      { id: 301, user_id: 1, collection_id: 3, name: 'Get User',    method: 'GET',    url: 'https://api.example.com/users/:id', headers: [], params: [{ k: 'id', v: '' }], body: '', auth: 'Bearer', token: '', created_at: '2026-02-18', updated_at: '2026-02-18' },
      { id: 302, user_id: 1, collection_id: 3, name: 'Update User', method: 'PUT',    url: 'https://api.example.com/users/:id', headers: [{ k: 'Content-Type', v: 'application/json' }], params: [], body: '{\n  "name": ""\n}', auth: 'Bearer', token: '', created_at: '2026-02-18', updated_at: '2026-02-18' },
      { id: 303, user_id: 1, collection_id: 3, name: 'Delete User', method: 'DELETE', url: 'https://api.example.com/users/:id', headers: [], params: [], body: '', auth: 'Bearer', token: '', created_at: '2026-02-18', updated_at: '2026-02-18' },
    ],
  },
  {
    id: 4, name: 'Webhook Listeners', tags: ['WebSocket', 'Events'],
    updated_at: '2026-02-15', default_method: 'WS', accent_color: '#f59e0b', pattern: 'lines',
    requests: [
      { id: 401, user_id: 1, collection_id: 4, name: 'Connect WS',       method: 'WS',   url: 'wss://api.example.com/events',     headers: [], params: [], body: '', auth: 'No Auth', token: '', created_at: '2026-02-15', updated_at: '2026-02-15' },
      { id: 402, user_id: 1, collection_id: 4, name: 'Subscribe Events', method: 'POST', url: 'https://api.example.com/webhooks', headers: [], params: [], body: '{\n  "events": ["payment.succeeded"]\n}', auth: 'Bearer', token: '', created_at: '2026-02-15', updated_at: '2026-02-15' },
    ],
  },
  {
    id: 5, name: 'Search Endpoints', tags: ['REST', 'Elastic'],
    updated_at: '2026-02-12', default_method: 'GET', accent_color: '#ec4899', pattern: 'cross',
    requests: [
      { id: 501, user_id: 1, collection_id: 5, name: 'Full-text Search', method: 'GET', url: 'https://api.example.com/search',         headers: [], params: [{ k: 'q', v: '' }, { k: 'page', v: '1' }], body: '', auth: 'Bearer', token: '', created_at: '2026-02-12', updated_at: '2026-02-12' },
      { id: 502, user_id: 1, collection_id: 5, name: 'Suggest',          method: 'GET', url: 'https://api.example.com/search/suggest', headers: [], params: [{ k: 'q', v: '' }], body: '', auth: 'Bearer', token: '', created_at: '2026-02-12', updated_at: '2026-02-12' },
    ],
  },
  {
    id: 6, name: 'Notification Service', tags: ['REST', 'Firebase'],
    updated_at: '2026-02-10', default_method: 'POST', accent_color: '#6366f1', pattern: 'waves',
    requests: [
      { id: 601, user_id: 1, collection_id: 6, name: 'Send Push',  method: 'POST', url: 'https://api.example.com/notifications/push',  headers: [{ k: 'Content-Type', v: 'application/json' }], params: [], body: '{\n  "token": "",\n  "title": "",\n  "body": ""\n}', auth: 'Bearer', token: '', created_at: '2026-02-10', updated_at: '2026-02-10' },
      { id: 602, user_id: 1, collection_id: 6, name: 'Send Email', method: 'POST', url: 'https://api.example.com/notifications/email', headers: [], params: [], body: '{\n  "to": "",\n  "subject": ""\n}', auth: 'Bearer', token: '', created_at: '2026-02-10', updated_at: '2026-02-10' },
    ],
  },
];
