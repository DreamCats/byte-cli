export class AuthError extends Error {
  constructor(message = "认证失败") {
    super(message);
    this.name = "AuthError";
  }
}

export class CookieNotFoundError extends AuthError {
  constructor(region: string) {
    super(`区域 ${region} 未配置 Cookie，请先运行 byte-cli auth config set-cookie`);
    this.name = "CookieNotFoundError";
  }
}

export class TokenFetchError extends AuthError {
  constructor(message: string) {
    super(`获取 Token 失败: ${message}`);
    this.name = "TokenFetchError";
  }
}

export class InvalidResponseError extends AuthError {
  constructor(message: string) {
    super(`无效的响应: ${message}`);
    this.name = "InvalidResponseError";
  }
}
