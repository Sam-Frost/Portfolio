export const LABEL_COLORS = ["red", "orange", "yellow", "green", "teal", "blue", "purple", "pink"] as const;

export type LabelColor = (typeof LABEL_COLORS)[number];

export interface Label {
  id: string;
  name: string;
  color: LabelColor;
}
