import { create } from "zustand";

interface UIStore {
  // Generic modal state
  openDialog: string | null;
  selectedId: string | null;
  setOpenDialog: (dialog: string | null, id?: string | null) => void;
  closeDialog: () => void;
}

export const useUIStore = create<UIStore>((set) => ({
  openDialog: null,
  selectedId: null,
  setOpenDialog: (dialog, id = null) =>
    set({ openDialog: dialog, selectedId: id }),
  closeDialog: () => set({ openDialog: null, selectedId: null }),
}));
