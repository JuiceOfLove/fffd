import $api from "../http";

export default class SubscriptionService {
  static async checkSubscription(): Promise<{ isActive: boolean }> {
    const res = await $api.get<{ isActive: boolean }>("/subscription/check");
    return res.data;
  }

  static async buySubscription(): Promise<{ payment_url: string }> {
    const res = await $api.post<{ payment_url: string }>("/subscription/buy");
    return res.data;
  }
}