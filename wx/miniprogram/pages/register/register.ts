type State = "UNSUBMITTED" | "PENDING" | "VERIFIED";

Page({
  /* 页面状态 */
  redirectURL: "",
  profileRefresher: 0,
  /* 视图数据 */
  data: {
    licenseState: "UNSUBMITTED" as State,
    licenseImgUrl: "",
    genders: ["未知", "男", "女"],
    form: {
      licenseId: "",
      name: "",
      gender: 0,
      birthDate: "1997-01-01",
    },
  },
  /* 生命周期函数 */
  onLoad(opt: Record<"redirect", string>) {
    const { redirect } = opt;
    if (redirect) {
      this.redirectURL = decodeURIComponent(redirect);
    }
  },
  onUnload() {},
  /* 页面方法 */
  changeGender(evt: any) {
    const { form } = this.data;
    form.gender = parseInt(evt.detail.value);
    this.setData({ form });
  },
  changeBirthDate(evt: any) {
    const { form } = this.data;
    form.birthDate = evt.detail.value;
    this.setData({ form });
  },
  submit() {},
  resubmit() {},
  uploadLicense() {
    wx.chooseMedia({
      success: async ({ tempFiles }) => {
        if (tempFiles.length === 0) return;

        this.setData({
          licenseImgUrl: tempFiles[0].tempFilePath,
        });
      },
    });
  },
  /* 辅助方法 */
  afterVerified() {
    if (this.redirectURL) {
      wx.redirectTo({
        url: this.redirectURL,
      });
    }
  },
});
