/// <reference types="cypress" />

describe('认证流程', () => {
  const testUser = {
    username: 'testuser',
    email: 'testuser@example.com',
    password: 'TestPassword123',
    name: '测试用户',
  };

  beforeEach(() => {
    cy.request({
      method: 'POST',
      url: '/api/v1/auth/register',
      body: testUser,
      failOnStatusCode: false // 已存在时忽略错误
    });
    cy.visit('/login');
    cy.clearLocalStorage();
  });

  it('登录失败时显示错误提示', () => {
    cy.get('input[name="email"]').type('wronguser@example.com');
    cy.get('input[name="password"]').type('wrongpassword');
    cy.get('button[type="submit"]').click();
    cy.contains('登录失败').should('be.visible');
  });

  it('登录成功后跳转首页并持久化状态', () => {
    cy.get('input[name="email"]').type(testUser.email);
    cy.get('input[name="password"]').type(testUser.password);
    cy.get('button[type="submit"]').click();
    cy.url().should('not.include', '/login');
    cy.window().then((win) => {
      expect(win.localStorage.getItem('auth_state')).to.exist;
    });
  });

  it('刷新页面后认证状态自动恢复', () => {
    cy.get('input[name="email"]').type(testUser.email);
    cy.get('input[name="password"]').type(testUser.password);
    cy.get('button[type="submit"]').click();
    cy.url().should('not.include', '/login');
    cy.reload();
    cy.url().should('not.include', '/login');
  });

  it('点击登出后清除状态并跳转登录页', () => {
    cy.get('input[name="email"]').type(testUser.email);
    cy.get('input[name="password"]').type(testUser.password);
    cy.get('button[type="submit"]').click();
    cy.url().should('not.include', '/login');
    cy.contains('登出').click();
    cy.url().should('include', '/login');
    cy.window().then((win) => {
      expect(win.localStorage.getItem('auth_state')).to.be.null;
    });
  });
}); 