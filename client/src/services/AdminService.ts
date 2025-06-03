import $api from "../http";

export default class AdminService {
  static async getPaymentHistory() {
    const response = await $api.get("/admin/payments");
    return response.data;
  }
}