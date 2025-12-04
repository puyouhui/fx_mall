class Employee {
  final int id;
  final String employeeCode;
  final String phone;
  final String name;
  final bool isDelivery;
  final bool isSales;
  final bool status;

  Employee({
    required this.id,
    required this.employeeCode,
    required this.phone,
    required this.name,
    required this.isDelivery,
    required this.isSales,
    required this.status,
  });

  factory Employee.fromJson(Map<String, dynamic> json) {
    return Employee(
      id: json['id'] as int,
      employeeCode: json['employee_code'] as String? ?? '',
      phone: json['phone'] as String,
      name: json['name'] as String,
      isDelivery: json['is_delivery'] as bool? ?? false,
      isSales: json['is_sales'] as bool? ?? false,
      status: json['status'] as bool? ?? true,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'employee_code': employeeCode,
      'phone': phone,
      'name': name,
      'is_delivery': isDelivery,
      'is_sales': isSales,
      'status': status,
    };
  }
}


