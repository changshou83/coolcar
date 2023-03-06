import { ModalResult } from "../../types/index";

Component({
  // 属性
  properties: {
    showModal: {
      type: Boolean,
      value: false,
    },
    showCancel: {
      type: Boolean,
      value: true,
    },
    title: {
      type: String,
      value: "弹窗",
    },
    confirmText: {
      type: String,
      value: "确认",
    },
    cancelText: {
      type: String,
      value: "取消",
    },
  },
  options: {
    addGlobalClass: true, // 组件样式隔离
  },
  // 视图数据
  data: {
    resolve: undefined as ((res: ModalResult) => void) | undefined,
  },
  // 视图方法
  methods: {
    cancel() {
      this.hideModal("cancel" as ModalResult);
    },
    confirm() {
      this.hideModal("confirm" as ModalResult);
    },
    hideModal(res: ModalResult) {
      this.setData({ showModal: false });

      this.triggerEvent(res);
      this.data.resolve?.(res);
    },
    showModal() {
      this.setData({ showModal: true });

      return new Promise((resolve) => {
        this.data.resolve = resolve;
      });
    },
  },
});
