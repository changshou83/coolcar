import { UserInfoKey } from "../constants/index";
import type { UserInfo } from "../types";

export const setUserInfo = (userInfo: UserInfo) => {
  wx.setStorageSync(UserInfoKey, userInfo);
};

export const getUserInfo = () => {
  return wx.getStorageSync(UserInfoKey) || {};
};
