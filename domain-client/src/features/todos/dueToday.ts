// Todo.targetDate is a plain "YYYY-MM-DD" date string from the backend with
// no timezone of its own (see server/internal/todo/model.go). Comparing it
// against "today" must use a fixed, deliberate calendar (IST) rather than
// the browser's local timezone, since a visitor's device clock has nothing
// to do with which calendar day the backend means.
const IST_TODAY_FORMATTER = new Intl.DateTimeFormat("en-CA", {
  timeZone: "Asia/Kolkata",
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
});

/** Today's calendar date in IST (Asia/Kolkata), as "YYYY-MM-DD". */
export function getTodayIST(): string {
  return IST_TODAY_FORMATTER.format(new Date());
}
