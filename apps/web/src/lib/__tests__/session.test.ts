import { beforeEach, describe, expect, it } from 'vitest';
import { clearSession, getSession, setSession } from '../session';

const uid = '11111111-2222-3333-4444-555555555555';

describe('session storage', () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it('returns null when nothing is stored', () => {
    expect(getSession()).toBeNull();
  });

  it('round-trips userId, username and name', () => {
    setSession({ userId: uid, username: 'joao', name: 'Sítio do João' });
    expect(getSession()).toEqual({ userId: uid, username: 'joao', name: 'Sítio do João' });
  });

  it('clears the session', () => {
    setSession({ userId: uid, username: 'joao', name: 'Sítio' });
    clearSession();
    expect(getSession()).toBeNull();
  });

  it('ignores corrupted storage', () => {
    window.localStorage.setItem('todo.session.v1', 'not json');
    expect(getSession()).toBeNull();
  });

  it('ignores entries missing userId or username', () => {
    window.localStorage.setItem('todo.session.v1', JSON.stringify({ username: 'maria' }));
    expect(getSession()).toBeNull();
  });

  it('defaults name to empty string when missing', () => {
    window.localStorage.setItem(
      'todo.session.v1',
      JSON.stringify({ userId: uid, username: 'maria' }),
    );
    expect(getSession()).toEqual({ userId: uid, username: 'maria', name: '' });
  });
});
