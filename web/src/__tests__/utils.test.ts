import { formatDate, formatDuration, cn } from '@/lib/utils';

describe('Utils', () => {
  describe('formatDate', () => {
    it('should format date string correctly', () => {
      const date = '2025-01-28T12:00:00Z';
      const formatted = formatDate(date);
      expect(formatted).toContain('Jan');
      expect(formatted).toContain('28');
      expect(formatted).toContain('2025');
    });

    it('should format Date object correctly', () => {
      const date = new Date('2025-01-28T12:00:00Z');
      const formatted = formatDate(date);
      expect(formatted).toContain('Jan');
      expect(formatted).toContain('28');
    });
  });

  describe('formatDuration', () => {
    it('should format milliseconds correctly', () => {
      expect(formatDuration(500)).toBe('500ms');
      expect(formatDuration(1500)).toBe('1.5s');
      expect(formatDuration(60000)).toBe('1.0m');
    });
  });

  describe('cn', () => {
    it('should join class names correctly', () => {
      expect(cn('foo', 'bar')).toBe('foo bar');
      expect(cn('foo', false, 'bar')).toBe('foo bar');
      expect(cn('foo', null, undefined, 'bar')).toBe('foo bar');
    });
  });
});
