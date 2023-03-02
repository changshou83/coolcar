import camelcaseKeys from "camelcase-keys";
import { baseURL } from "../constants/index";
import { getAuthData, login, setAuthData } from "./index";

const defaultAuthOpts = {
  attachAuthHeader: true, // 是否携带token
  retryOnAuthError: true, // 是否在认证失败时重试
};
const AUTH_ERR = "AUTH_ERR";

export interface RequestOption<REQ, RES> {
  method: "GET" | "PUT" | "POST" | "DELETE";
  url: string;
  data?: REQ;
  resolveRes: (r: object) => RES;
}

export interface AuthData {
  token: string; // token
  expiryMs: number; // 什么时候过期
}

export interface AuthOption {
  attachAuthHeader: boolean;
  retryOnAuthError: boolean;
}

export async function requestWithRetry<
  REQ extends WechatMiniprogram.IAnyObject,
  RES
>(
  opts: RequestOption<REQ, RES>,
  authOpt: AuthOption = defaultAuthOpts
): Promise<RES> {
  try {
    return await request(opts, authOpt);
  } catch (err) {
    if (err === AUTH_ERR && authOpt.retryOnAuthError) {
      // 如果token过期并且没重试过就刷新token
      setAuthData({
        token: "",
        expiryMs: 0,
      });
      await login();
      // 然后重试,PS：小心闭包缓存 retryOnAuthError
      return requestWithRetry(opts, {
        ...authOpt,
        retryOnAuthError: false,
      });
    } else {
      throw err;
    }
  }
}

export function request<REQ extends WechatMiniprogram.IAnyObject, RES>(
  opts: RequestOption<REQ, RES>,
  authOpt: AuthOption
): Promise<RES> {
  return new Promise((resolve, reject) => {
    const header: Record<string, any> = {};
    if (authOpt.attachAuthHeader) {
      const { token, expiryMs } = getAuthData();
      console.log({
        token,
        noExpire: expiryMs >= Date.now(),
      });
      if (token && expiryMs >= Date.now()) {
        header.authorization = "Bearer " + token;
      } else {
        reject(AUTH_ERR);
        return;
      }
    }

    const { url, method, data } = opts;
    wx.request<REQ>({
      header,
      method,
      data,
      url: baseURL + url,
      success: (res) => {
        const { statusCode: code, data } = res;
        if (code >= 400) {
          reject(code === 401 ? AUTH_ERR : res);
        } else {
          const res = camelcaseKeys(data as object, { deep: true });
          resolve(opts.resolveRes(res));
        }
      },
      fail: reject,
    });
  });
}
