export const getSetting = (): Promise<WechatMiniprogram.GetSettingSuccessCallbackResult> => new Promise((resolve, reject) => {
  wx.getSetting({
    success: resolve,
    fail: reject,
  })
})

export const getUserInfo = (): Promise<WechatMiniprogram.GetUserInfoSuccessCallbackResult> => new Promise((resolve, reject) => {
  wx.getUserInfo({
    success: resolve,
    fail: reject,
  })
})
