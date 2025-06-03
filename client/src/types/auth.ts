export interface IUser {
    id: number;
    name: string;
    email: string;
    role: string;
    family_id: number;
    isActivated: boolean; // добавляем свойство isActivated
    activationLink?: string;
    created_at: string;
    updated_at: string;
  }

  export interface IAuthResponse {
    access_token: string;
    refresh_token: string;
    user: IUser;
  }

  export interface ILoginRequest {
    email: string;
    password: string;
  }

  export interface IRegistrationRequest {
    name: string;
    email: string;
    password: string;
  }