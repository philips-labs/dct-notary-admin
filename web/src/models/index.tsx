export interface Target {
  id: string;
  gun?: string;
  role?: string;
}

export interface Delegation {
  id: string;
  role: string;
}

export interface TargetListData {
  targets: Target[];
}

export interface DelegationListData {
  delegations: Delegation[];
}
