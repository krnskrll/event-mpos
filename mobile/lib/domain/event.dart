enum EventStatus { draft, active, closed }

class AssignedEvent {
  const AssignedEvent({
    required this.id,
    required this.code,
    required this.name,
    this.startDate,
    this.endDate,
    required this.status,
  });

  factory AssignedEvent.fromJson(Map<String, dynamic> json) => AssignedEvent(
    id: json['id'],
    code: json['code'],
    name: json['name'],
    startDate: json['startDate'],
    endDate: json['endDate'],
    status: EventStatus.values.asNameMap()[(json['status'] as String?)?.toLowerCase()] ?? EventStatus.draft,
  );

  final String id;
  final String code;
  final String name;
  final String? startDate;
  final String? endDate;
  final EventStatus status;
}
