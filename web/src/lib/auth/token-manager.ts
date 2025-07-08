import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import axios from 'axios';
import { writable, type Writable } from 'svelte/store';

interface TokenData {
	accessToken: string;
	refreshToken: string;
	expiresIn: number;
}

interface RefreshResponse {
	access_token: string;
	refresh_token: string;
	expires_in: number;
}

export class TokenManager {
	private static instance: TokenManager;
	private refreshPromise: Promise<void> | null = null;
	private refreshTimeout: ReturnType<typeof setTimeout> | null = null;
	private tokenStore: Writable<TokenData | null>;
	private readonly storageKey = 'stellar_auth_tokens';

	private constructor() {
		this.tokenStore = writable(this.loadTokens());
		this.setupRefreshTimer();
	}

	static getInstance(): TokenManager {
		if (!TokenManager.instance) {
			TokenManager.instance = new TokenManager();
		}
		return TokenManager.instance;
	}

	private loadTokens(): TokenData | null {
		if (!browser) return null;
		const stored = localStorage.getItem(this.storageKey);
		if (!stored) return null;
		try {
			return JSON.parse(stored);
		} catch {
			return null;
		}
	}

	private saveTokens(tokens: TokenData | null): void {
		if (!browser) return;
		if (tokens) {
			localStorage.setItem(this.storageKey, JSON.stringify(tokens));
		} else {
			localStorage.removeItem(this.storageKey);
		}
		this.tokenStore.set(tokens);
	}

	private setupRefreshTimer(): void {
		const tokens = this.loadTokens();
		if (!tokens) return;

		const expiresIn = tokens.expiresIn * 1000; // 转换为毫秒
		const refreshBuffer = 5 * 60 * 1000; // 提前5分钟刷新
		const timeUntilRefresh = Math.max(0, expiresIn - refreshBuffer);

		if (this.refreshTimeout) {
			clearTimeout(this.refreshTimeout);
		}

		this.refreshTimeout = setTimeout(() => {
			this.refreshTokens().catch(console.error);
		}, timeUntilRefresh);
	}

	async refreshTokens(): Promise<void> {
		const tokens = this.loadTokens();
		if (!tokens?.refreshToken) {
			this.handleTokenExpired();
			return;
		}

		// 防止并发刷新
		if (this.refreshPromise) {
			return this.refreshPromise;
		}

		this.refreshPromise = new Promise<void>(async (resolve, reject) => {
			try {
				const response = await axios.post<RefreshResponse>('/api/v1/auth/refresh', {
					refresh_token: tokens.refreshToken
				});

				const newTokens: TokenData = {
					accessToken: response.data.access_token,
					refreshToken: response.data.refresh_token,
					expiresIn: response.data.expires_in
				};

				this.saveTokens(newTokens);
				this.setupRefreshTimer();
				resolve();
			} catch (error) {
				this.handleTokenExpired();
				reject(error);
			} finally {
				this.refreshPromise = null;
			}
		});

		return this.refreshPromise;
	}

	private handleTokenExpired(): void {
		this.saveTokens(null);
		if (browser) {
			goto('/login');
		}
	}

	getAccessToken(): string | null {
		const tokens = this.loadTokens();
		return tokens?.accessToken || null;
	}

	setTokens(tokens: TokenData): void {
		this.saveTokens(tokens);
		this.setupRefreshTimer();
	}

	clearTokens(): void {
		if (this.refreshTimeout) {
			clearTimeout(this.refreshTimeout);
		}
		this.saveTokens(null);
	}

	subscribe(callback: (tokens: TokenData | null) => void) {
		return this.tokenStore.subscribe(callback);
	}
}

export const tokenManager = TokenManager.getInstance();
