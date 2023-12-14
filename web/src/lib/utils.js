export function formatDateToYYYYMMDD(dateInput) {
  const date = new Date(dateInput);
  let year = date.getFullYear();
  let month = date.getMonth() + 1; // getMonth() returns a zero-based index
  let day = date.getDate();

  // Pad the month and day with leading zeros if necessary
  month = month < 10 ? '0' + month : month;
  day = day < 10 ? '0' + day : day;

  return `${year}.${month}.${day}`;
}