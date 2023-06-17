import { auth } from "../apis/proto_gen/auth/auth_pb";
import { getAuthData, request, setAuthData } from "./index";

export async function login() {
  const { token, expiryMs } = getAuthData();
  if (token && expiryMs >= Date.now()) {
    return;
  }
  const { code } = await wxLogin();
  const reqTimeMs = Date.now();
  const { accessToken, expiresIn } = await request<
    auth.v1.ILoginRequest,
    auth.v1.ILoginResponse
  >(
    {
      method: "POST",
      url: "/v1/auth/login",
      data: { code },
      resolveRes: auth.v1.LoginResponse.fromObject,
    },
    {
      attachAuthHeader: false,
      retryOnAuthError: false,
    }
  );
  wx.showToast({ title: "登录成功", icon: "success", mask: false });
  setAuthData({
    token: accessToken!,
    expiryMs: reqTimeMs + expiresIn! * 1000,
  });
}

function wxLogin(): Promise<WechatMiniprogram.LoginSuccessCallbackResult> {
  return new Promise((resolve, reject) => {
    wx.login({
      success: resolve,
      fail: reject,
    });
  });
}
