export default function SortQuery(s) {
  let query = "";
  for (let i = 0; i < s.length; i++) {
    const e = s[i];

    query += `&orderBy=${e.key},${e.order.toUpperCase()}`;
  }
  return query;
}
