enum EventStatus { draft, active, closed }

/// The location this device sells at: an event (temporary) or an outlet (permanent).
class AssignedEvent {
  const AssignedEvent({
    required this.id,
    required this.code,
    required this.name,
    this.startDate,
    this.endDate,
    required this.status,
    this.isOutlet = false,
  });

  factory AssignedEvent.fromJson(Map<String, dynamic> json) => AssignedEvent(
    id: json['id'],
    code: json['code'] ?? '',
    name: json['name'],
    startDate: json['startDate'],
    endDate: json['endDate'],
    status: EventStatus.values.asNameMap()[(json['status'] as String?)?.toLowerCase()] ?? EventStatus.draft,
    isOutlet: json['type'] == 'outlet',
  );

  Map<String, dynamic> toJson() => {
    'id': id,
    'code': code,
    'name': name,
    'startDate': ?startDate,
    'endDate': ?endDate,
    'status': status.name.toUpperCase(),
    'type': isOutlet ? 'outlet' : 'event',
  };

  final String id;
  final String code;
  final String name;
  final String? startDate;
  final String? endDate;
  final EventStatus status;
  final bool isOutlet;
}

const _months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

/// Plain-words dates for event cards and Home.
extension EventDates on AssignedEvent {
  DateTime? get _start => DateTime.tryParse(startDate ?? '');
  DateTime? get _end => DateTime.tryParse(endDate ?? '');

  /// "1 – 14 Jul 2026", "28 Jun – 3 Jul 2026" or "30 Dec 2026 – 2 Jan 2027".
  String get dateRange {
    final (s, e) = (_start, _end);
    if (s == null || e == null) return [?startDate, ?endDate].join(' – ');
    final end = '${e.day} ${_months[e.month - 1]} ${e.year}';
    if (s.year != e.year) return '${s.day} ${_months[s.month - 1]} ${s.year} – $end';
    if (s.month != e.month) return '${s.day} ${_months[s.month - 1]} – $end';
    return '${s.day} – $end';
  }

  /// "Outlet", or "Event · 10 – 20 Mar 2027".
  String get placeLabel => isOutlet ? 'Outlet' : ['Event', if (dateRange.isNotEmpty) dateRange].join(' · ');

  /// "Day 3 of 14" while the event is running on [today].
  String? dayOf(DateTime today) {
    final (s, e, t) = (_start, _end, DateTime(today.year, today.month, today.day));
    if (s == null || e == null || t.isBefore(s) || t.isAfter(e)) return null;
    return 'Day ${t.difference(s).inDays + 1} of ${e.difference(s).inDays + 1}';
  }

  /// "Opens today", "Opens tomorrow" or "Opens in 18 days" before the start date.
  String? opensIn(DateTime today) {
    final s = _start;
    final t = DateTime(today.year, today.month, today.day);
    if (s == null || t.isAfter(s)) return null;
    return switch (s.difference(t).inDays) {
      0 => 'Opens today',
      1 => 'Opens tomorrow',
      final n => 'Opens in $n days',
    };
  }

  /// How far through the event [today] is, 0–1; null outside its dates.
  double? progress(DateTime today) {
    final (s, e, t) = (_start, _end, DateTime(today.year, today.month, today.day));
    if (s == null || e == null || t.isBefore(s) || t.isAfter(e)) return null;
    return (t.difference(s).inDays + 1) / (e.difference(s).inDays + 1);
  }
}
