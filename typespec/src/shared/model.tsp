namespace Shared;

model ProjectDynamicField {
    label: string;
    name: string;
    type: string;
    isRequired: boolean;
    defaultValue?: string;
}

model AssetInfo {
    @key id: string;
    name: string;
}

model Asset {
    ...AssetInfo;
    size: int64;
    isEnabled: boolean;
    lastModified: utcDateTime;
}
