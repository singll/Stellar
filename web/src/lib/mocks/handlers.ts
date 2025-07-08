import { http, HttpResponse } from 'msw';

// 模拟用户数据
const mockUsers = [
	{
		id: '1',
		username: 'admin',
		email: 'admin@stellar.com',
		role: 'admin',
		name: '管理员'
	}
];

// 模拟 token
const generateToken = (userId: string) => `mock-token-${userId}-${Date.now()}`;

export const handlers = [
	// 登录接口
	http.post('/api/v1/auth/login', async ({ request }) => {
		const { username, password } = await request.json();

		// 模拟验证
		const user = mockUsers.find((u) => u.username === username);

		if (!user || password !== 'admin123') {
			return new HttpResponse(
				JSON.stringify({
					code: 401,
					message: '用户名或密码错误'
				}),
				{ status: 401 }
			);
		}

		return HttpResponse.json({
			code: 200,
			message: 'success',
			data: {
				token: generateToken(user.id),
				user
			}
		});
	}),

	// 获取当前用户信息
	http.get('/api/v1/auth/me', ({ request }) => {
		const authHeader = request.headers.get('Authorization');

		if (!authHeader?.startsWith('Bearer ')) {
			return new HttpResponse(null, { status: 401 });
		}

		return HttpResponse.json({
			code: 200,
			message: 'success',
			data: mockUsers[0]
		});
	}),

	// 登出接口
	http.post('/api/v1/auth/logout', () => {
		return HttpResponse.json({
			code: 200,
			message: 'success'
		});
	})
];
