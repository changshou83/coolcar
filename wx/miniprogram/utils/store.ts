import { AuthDataKey, UserInfoKey } from "../constants/index";
import type { UserInfo } from "../types";
import { AuthData } from "./request";

export const setUserInfo = (userInfo: UserInfo) => {
  wx.setStorageSync(UserInfoKey, userInfo);
};

export const getUserInfo = () => {
  return wx.getStorageSync(UserInfoKey) || {};
};

export const setAuthData = (authData: AuthData) => {
  wx.setStorageSync(AuthDataKey, authData);
};

export const getAuthData = (): AuthData => {
  return (
    wx.getStorageSync(AuthDataKey) || {
      token: "",
      expiryMs: 0,
    }
  );
};
