import { defaultAvatar } from "../../constants/index";
import { getUserInfo, routing, setUserInfo } from "../../utils/index";

const enableShareKey = "enable_share_location";

Page({
  /* 页面状态 */
  carID: "",
  carRefresher: 0,
  /* 页面数据 */
  data: {
    avatarUrl: defaultAvatar,
    enableShareLoc: false,
  },
  /* 生命周期函数 */
  async onLoad(opts: Record<"car_id", string>) {
    const { car_id } = opts as routing.LockOpts;
    this.carID = car_id;

    const { avatarURL } = getUserInfo();
    this.setData({
      avatarUrl: avatarURL || defaultAvatar,
      enableShareLoc: wx.getStorageSync(enableShareKey) || false,
    });
  },
  onUnload() {
    this.clearCarRefresher();
    wx.hideLoading();
  },
  /* 页面方法 */
  shareLocation(evt: any) {
    const { value } = evt.detail;
    this.data.enableShareLoc = value; // 不用setData进行重新渲染
    wx.setStorageSync(enableShareKey, value);
  },
  unlock() {
    wx.getLocation({
      type: "gcj02",
      success: async () => {
        // if (!this.carID) {
        //   console.error("no car id specified");
        //   return;
        // }
        wx.showToast({
          title: "开锁中...",
          mask: true,
        });
        this.carRefresher = setInterval(async () => {
          const state = "UNLOCKED";
          const trip_id = "1";
          if (state === "UNLOCKED") {
            this.clearCarRefresher();
            wx.redirectTo({
              url: routing.driving({ trip_id }),
              complete() {
                wx.hideLoading();
              },
            });
          }
        }, 2000);
      },
      fail() {
        wx.showToast({
          icon: "none",
          title: "请前往设置页授权位置信息",
        });
      },
    });
  },
  chooseAvatar(evt: any) {
    const { avatarUrl } = evt.detail;
    const userInfo = getUserInfo();
    if (avatarUrl) {
      this.setData({ avatarUrl });
      setUserInfo({ ...userInfo, avatarURL: avatarUrl });
    }
  },
  /* 辅助方法 */
  clearCarRefresher() {
    if (this.carRefresher) {
      clearInterval(this.carRefresher);
      this.carRefresher = 0;
    }
  },
});
