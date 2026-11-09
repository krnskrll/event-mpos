/// A cashier or supervisor of this device's location, as cached for signing in offline.
class StaffMember {
  const StaffMember({required this.id, required this.name, required this.isSupervisor, required this.pinHash});

  final String id;
  final String name;
  final bool isSupervisor;

  /// bcrypt hash of the 4–6 digit PIN, checked on the device.
  final String pinHash;

  String get roleLabel => isSupervisor ? 'Supervisor' : 'Cashier';
}
