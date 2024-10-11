export const UsersTableHeaders = [
  { key: "user_id", title: "ID" },
  { key: "name", title: "Name" },
  { key: "email", title: "Email" },
  { key: "roles", title: "Roles" },
  { key: "department", title: "Department", sortable: false },
  { key: "enabled", title: "Enabled" },
  { key: "dateCreated", title: "Date Created", sortable: false },
  { key: "dateUpdated", title: "Date Updated", sortable: false },
  { key: "actions", title: "Actions", sortable: false },
];

export const UserHomesTableHeaders = [
  { key: "type", title: "Type" },
  { key: "address.address1", title: "Address First Line", sortable: false },
  { key: "address.address2", title: "Address Second Line", sortable: false },
  { key: "address.zipCode", title: "ZIP Code", sortable: false },
  { key: "address.city", title: "City", sortable: false },
  { key: "address.state", title: "State", sortable: false },
  { key: "address.country", title: "Country", sortable: false },
  { key: "actions", title: "Actions", sortable: false },
];
