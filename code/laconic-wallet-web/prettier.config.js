const config = {
  arrowParens: "always",
  printWidth: 80,
  semi: false,
  singleQuote: true,
  tabWidth: 2,
  trailingComma: "es5",
  importOrderSeparation: true,
  importOrderSortSpecifiers: true,
  plugins: ["@trivago/prettier-plugin-sort-imports"],
  importOrder: [
    "<THIRD_PARTY_MODULES>",
    "^(pages|components|utils|icons|test|graphql)/(.*)$",
    "^[./]",
  ],
};

export default config;
