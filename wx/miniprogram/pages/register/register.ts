import { submitProfile, getProfile, clearProfile } from "../../apis/profile";
import { rental } from "../../apis/proto_gen/rental/rental_pb";
import { formatDate } from "../../utils/index";

const State = rental.v1.IdentityStatus;

Page({
  /* 页面状态 */
  redirectURL: "",
  profileRefresher: 0,
  /* 视图数据 */
  data: {
    licenseState: State[State.UNSUBMITTED],
    licenseImgUrl: "",
    genders: ["未知", "男", "女"],
    form: {
      licNumber: "",
      name: "",
      gender: 0,
      birthDate: "1997-01-01",
    },
  },
  /* 生命周期函数 */
  async onLoad(opt: Record<"redirect", string>) {
    const { redirect } = opt;
    if (redirect) {
      this.redirectURL = decodeURIComponent(redirect);
    }
    // render profile
    const profile = await getProfile();
    this.renderProfile(profile);
  },
  onUnload() {
    this.clearProfileRefresher();
  },
  /* 页面方法 */
  changeLicNumber(evt: any) {
    this.changeForm("licNumber", evt.detail.value);
  },
  changeName(evt: any) {
    this.changeForm("name", evt.detail.value);
  },
  changeGender(evt: any) {
    this.changeForm("gender", evt.detail.value);
  },
  changeBirthDate(evt: any) {
    this.changeForm("birthDate", evt.detail.value);
  },
  async submit() {
    // create form
    const form: any = {
      ...this.data.form,
      birthDateMs: Date.parse(this.data.form.birthDate),
    };
    delete form.birthDate;
    // validate
    const identityReg =
      /^[1-9]\d{5}[1-9]\d{3}((0\d)|(1[0-2]))(([0-2]\d)|3[0-1])\d{3}(\d|x|X)$/;
    if (!identityReg.test(form.licNumber)) {
      wx.showToast({
        title: "驾驶证号不正确",
        icon: "error",
        duration: 1500,
      });
      return;
    } else if (form.name === "") {
      wx.showToast({
        title: "请填写姓名",
        icon: "error",
        duration: 1500,
      });
      return;
    }
    // submit
    const profile = await submitProfile(form);
    this.renderProfile(profile);
    this.scheduleProfileRefresher();
  },
  async resubmit() {
    const profile = await clearProfile();
    this.renderProfile(profile);
  },
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
  scheduleProfileRefresher() {
    // 轮询查看状态
    this.profileRefresher = setInterval(async () => {
      const profile = await getProfile();
      this.renderProfile(profile);

      profile.status !== State.PENDING && this.clearProfileRefresher();
      profile.status === State.VERIFIED && this.afterVerified();
    }, 1000);
  },
  clearProfileRefresher() {
    if (this.profileRefresher !== 0) {
      clearInterval(this.profileRefresher);
      this.profileRefresher = 0;
    }
  },
  getNewIdentity(identity?: rental.v1.IIdentity) {
    let form = this.data.form;
    form = {
      licNumber: identity?.licNumber || "",
      name: identity?.name || "",
      gender: identity?.gender || 0,
      birthDate: formatDate(identity?.birthDateMs || 0),
    };
    return form;
  },
  renderProfile(profile: rental.v1.IProfile) {
    const form = this.getNewIdentity(profile.identity!);
    this.setData({
      form,
      licenseState: State[profile.status || State.UNSUBMITTED],
    });
  },
  afterVerified() {
    if (this.redirectURL) {
      wx.redirectTo({ url: this.redirectURL });
    }
  },
  changeForm(k: "birthDate" | "licNumber" | "name" | "gender", newVal: string) {
    const { form } = this.data;
    if (k == "gender") {
      form[k] = parseInt(newVal);
    } else {
      form[k] = newVal;
    }
    this.setData({ form });
  },
});
