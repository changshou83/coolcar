interface Marker {
  iconPath: string;
  id: number;
  latitude: number;
  longitude: number;
  width: number;
  height: number;
}

const defaultAvatar = "/resources/car.png";
const initialLat = 42.05297;
const initialLng = 123.52658;

Page({
  /* 页面状态 */
  isFrontDesk: false,
  socket: undefined as WechatMiniprogram.SocketTask | undefined,
  /* 视图所用数据 */
  data: {
    avatarURL: "",
    scale: 16,
    location: {
      latitude: initialLat,
      longitude: initialLng,
    },
    setting: {
      skew: 0,
      rotate: 0,
      showLocation: true,
      showScale: true,
      subKey: "",
      layerStyle: -1,
      enableZoom: true,
      enableScroll: true,
      enableRotate: false,
      showCompass: false,
      enable3D: false,
      enableOverlooking: false,
      enableSatellite: false,
      enableTraffic: false,
    },
    markers: [] as Marker[],
  },
  /* 生命周期 */
  onLoad() {
    const {
      globalData: { userInfo },
    } = getApp<IAppOption>();
    this.setData({
      avatarURL: userInfo?.avatarUrl,
    });
  },
  onShow() {
    this.isFrontDesk = true;
    if (!this.socket) {
      this.setData({ markers: [] }, () => this.setupCarPosUpdater());
    }
  },
  onHide() {
    this.isFrontDesk = false;
    if (this.socket) {
      this.socket.close({
        success: () => (this.socket = undefined),
      });
    }
  },
  /* 页面方法 */
  // 跳转到我的行程
  gotoTrips() {
    // TODO
  },
  // 定位当前位置
  locateCurLoc() {
    wx.getLocation({
      type: "gcj02", // 返回可用于 wx.openLocation 的坐标
      success: (res) => {
        this.setData({
          location: {
            latitude: res.latitude,
            longitude: res.longitude,
          },
        });
      },
      fail() {
        // 提示用户
        wx.showToast({
          icon: "none",
          title: "请前往设置页授权",
        });
      },
    });
  },
  // 扫码
  async scanCode() {
    wx.scanCode({
      success: () => {
        wx.navigateTo({
          url: "/pages/register/register",
        });
      },
      fail: console.error,
    });
  },
  /* 辅助方法 */
  setupCarPosUpdater() {
    // get map context
    // const mapCtx = wx.createMapContext("map");
    // const markers = new Map<string, Marker>();
    // lock
    // let translating = false;
    // const endTranslation = () => translating = false;
    // const updateMarker = (fn: () => void) => {
    //   translating = true;
    //   fn()
    // }
    // get/create marker -> move marker
    // TODO: CarService
    // this.socket = CarService.subscribe(({ id, car }) => {
    //   if (!id || translating || !this.isFrontDesk) {
    //     console.log("dropped");
    //     return
    //   }
    //   const { driver, position } = car;
    //   const newIcon = driver.avatarUrl || defaultAvatar;
    //   const newLat = position.latitude || initialLat;
    //   const newLng = position.longitude || initialLng;
    //   const marker = markers.get(id);
    //   // create new marker
    //   if(!marker) {
    //     const { markers: _markers } = this.data;
    //     const newMarker: Marker = {
    //       id: _markers.length,
    //       iconPath: newIcon,
    //       latitude: newLat,
    //       longitude: newLng,
    //       height: 20,
    //       width: 20
    //     };
    //     // insert new marker
    //     markers.set(id, newMarker);
    //     _markers.push(newMarker);
    //     // update view
    //     updateMarker(() => this.setData({ markers: this.data.markers }, endTranslation))
    //     return
    //   }
    //   // Change Icon
    //   if (marker.iconPath !== newIcon) {
    //     marker.iconPath = newIcon;
    //     marker.latitude = newLat;
    //     marker.longitude = newLng;
    //     updateMarker(() => this.setData({ markers: this.data.markers }, endTranslation))
    //     return
    //   }
    //   // Move Marker
    //   if (marker.latitude !== newLat || marker.longitude !== newLng) {
    //     updateMarker(() => mapCtx.translateMarker({
    //       markerId: marker.id,
    //       destination: {
    //         latitude: newLat,
    //         longitude: newLng,
    //       },
    //       autoRotate: false,
    //       rotate: 0,
    //       duration: 80,
    //       animationEnd: endTranslation,
    //     }))
    //   }
    // })
  },
});
