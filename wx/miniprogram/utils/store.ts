import type { UserInfo } from "../types";

const UserInfoKey = "user_info";

export const setUserInfo = (userInfo: UserInfo) => {
  wx.setStorageSync(UserInfoKey, userInfo);
};

export const getUserInfo = () => {
  return wx.getStorageSync(UserInfoKey) || {};
};
