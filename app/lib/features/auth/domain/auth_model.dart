class AuthUser {
  final String id;
  final String email;
  final String nickname;
  final String role;
  final String locale;

  const AuthUser({
    required this.id,
    required this.email,
    required this.nickname,
    required this.role,
    required this.locale,
  });

  factory AuthUser.fromJson(Map<String, dynamic> json) {
    return AuthUser(
      id: json['id']?.toString() ?? '',
      email: json['email']?.toString() ?? '',
      nickname: json['nickname']?.toString() ?? '',
      role: json['role']?.toString() ?? 'user',
      locale: json['locale']?.toString() ?? 'en-US',
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'email': email,
    'nickname': nickname,
    'role': role,
    'locale': locale,
  };

  AuthUser copyWith({
    String? id,
    String? email,
    String? nickname,
    String? role,
    String? locale,
  }) {
    return AuthUser(
      id: id ?? this.id,
      email: email ?? this.email,
      nickname: nickname ?? this.nickname,
      role: role ?? this.role,
      locale: locale ?? this.locale,
    );
  }
}

class AuthTokens {
  final String accessToken;
  final String refreshToken;

  const AuthTokens({
    required this.accessToken,
    required this.refreshToken,
  });

  factory AuthTokens.fromJson(Map<String, dynamic> json) {
    return AuthTokens(
      accessToken: json['access_token']?.toString() ?? '',
      refreshToken: json['refresh_token']?.toString() ?? '',
    );
  }
}
