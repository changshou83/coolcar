export namespace routing {
  /* interfaces */
  export interface DrivingOpts {
    trip_id: string;
  }
  export interface LockOpts {
    car_id: string;
  }
  export interface RegisterOpts {
    redirect?: string;
  }
  export interface RegisterParams {
    redirectURL: string;
  }
  /* functions */
  export function driving(opts: DrivingOpts) {
    return `/pages/driving/driving?trip_id=${opts.trip_id}`;
  }
  export function lock(opts: LockOpts) {
    return `/pages/lock/lock?car_id=${opts.car_id}`;
  }
  export function register(params?: RegisterParams) {
    return (
      "/pages/register/register" +
      (params ? `?redirect=${encodeURIComponent(params.redirectURL)}` : "")
    );
  }
}
