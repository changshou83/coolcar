export interface UploadFileOpts {
  localPath: string;
  url: string;
}

export function uploadFile(opts: UploadFileOpts): Promise<void> {
  const data = wx.getFileSystemManager().readFileSync(opts.localPath);
  return new Promise((resolve, reject) => {
    wx.request({
      method: "PUT",
      url: opts.url,
      data,
      success(res) {
        res.statusCode >= 400 ? reject(res) : resolve();
      },
      fail: reject,
    });
  });
}
