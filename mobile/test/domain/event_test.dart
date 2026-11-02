import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/domain/event.dart';

AssignedEvent _event(String start, String end) =>
    AssignedEvent(id: 'e', code: 'EV', name: 'Fair', startDate: start, endDate: end, status: EventStatus.active);

void main() {
  test('date range collapses the shared month and year', () {
    expect(_event('2026-07-01', '2026-07-14').dateRange, '1 – 14 Jul 2026');
    expect(_event('2026-06-28', '2026-07-03').dateRange, '28 Jun – 3 Jul 2026');
    expect(_event('2026-12-30', '2027-01-02').dateRange, '30 Dec 2026 – 2 Jan 2027');
  });

  test('day of the event while it runs, opening countdown before it starts', () {
    final fair = _event('2026-11-20', '2026-11-23');
    expect(fair.dayOf(DateTime(2026, 11, 22, 15)), 'Day 3 of 4');
    expect(fair.dayOf(DateTime(2026, 11, 24)), isNull);
    expect(
      [fair.opensIn(DateTime(2026, 11, 2)), fair.opensIn(DateTime(2026, 11, 19)), fair.opensIn(DateTime(2026, 11, 20))],
      ['Opens in 18 days', 'Opens tomorrow', 'Opens today'],
    );
    expect(fair.opensIn(DateTime(2026, 11, 21)), isNull);
  });
}
