import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/fcm_service.dart';
import '../data/auth_repository.dart';
import '../domain/auth_model.dart';

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

class AuthState {
  final AuthUser? user;
  final bool isAuthenticated;
  final bool isLoading;
  final String? error;

  const AuthState({
    this.user,
    this.isAuthenticated = false,
    this.isLoading = false,
    this.error,
  });

  AuthState copyWith({
    AuthUser? user,
    bool? isAuthenticated,
    bool? isLoading,
    String? error,
    bool clearError = false,
    bool clearUser = false,
  }) {
    return AuthState(
      user: clearUser ? null : (user ?? this.user),
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
      isLoading: isLoading ?? this.isLoading,
      error: clearError ? null : (error ?? this.error),
    );
  }
}

// ---------------------------------------------------------------------------
// Notifier
// ---------------------------------------------------------------------------

class AuthNotifier extends StateNotifier<AuthState> {
  final AuthRepository _repository;
  final Ref _ref;

  AuthNotifier(this._repository, this._ref) : super(const AuthState()) {
    init();
  }

  /// Called at app start – checks whether a stored token/user exists.
  Future<void> init() async {
    state = state.copyWith(isLoading: true);
    try {
      final hasToken = await _repository.hasValidToken();
      if (hasToken) {
        final user = await _repository.getStoredUser();
        state = AuthState(
          user: user,
          isAuthenticated: true,
          isLoading: false,
        );
      } else {
        state = const AuthState(isLoading: false);
      }
    } catch (_) {
      state = const AuthState(isLoading: false);
    }
  }

  Future<void> login(String email, String password) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repository.login(email, password);
      final user = await _repository.getStoredUser();
      state = AuthState(
        user: user,
        isAuthenticated: true,
        isLoading: false,
      );
      // Notify GoRouter listeners
      _authNotifier.notifyListeners();
      // FCM 토큰 등록 (firebase_messaging 패키지 추가 후 활성화)
      FCMService.registerToken(_ref, user?.locale ?? 'ko').ignore();
    } on Exception catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: _friendlyMessage(e),
      );
    }
  }

  Future<void> register(
    String email,
    String password,
    String nickname,
    String locale,
  ) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repository.register(email, password, nickname, locale);
      state = state.copyWith(isLoading: false);
    } on Exception catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: _friendlyMessage(e),
      );
    }
  }

  Future<void> logout() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repository.logout();
    } finally {
      state = const AuthState(isLoading: false);
      _authNotifier.notifyListeners();
    }
  }

  String _friendlyMessage(Exception e) {
    final msg = e.toString();
    if (msg.contains('401') || msg.contains('Unauthorized')) {
      return 'Invalid email or password.';
    }
    if (msg.contains('409') || msg.contains('Conflict')) {
      return 'An account with this email already exists.';
    }
    if (msg.contains('SocketException') || msg.contains('connection')) {
      return 'No internet connection. Please try again.';
    }
    return 'Something went wrong. Please try again.';
  }
}

// ---------------------------------------------------------------------------
// ChangeNotifier bridge for GoRouter refreshListenable
// ---------------------------------------------------------------------------

/// A simple ChangeNotifier that GoRouter can listen to for auth state changes.
/// AuthNotifier calls notifyListeners() on login/logout.
class AuthChangeNotifier extends ChangeNotifier {
  // Public so AuthNotifier can call it
  @override
  void notifyListeners() => super.notifyListeners();
}

/// Singleton instance used by both GoRouter and AuthNotifier.
final _authNotifier = AuthChangeNotifier();

final authChangeNotifierProvider = Provider<AuthChangeNotifier>((_) => _authNotifier);

// ---------------------------------------------------------------------------
// Providers
// ---------------------------------------------------------------------------

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  final repository = ref.watch(authRepositoryProvider);
  return AuthNotifier(repository, ref);
});
