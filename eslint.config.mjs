import stylistic from "@stylistic/eslint-plugin";
import vue from "eslint-plugin-vue";

export default [
  {
    ignores: [
      "backend/**",
      "node_modules/**",
      "frontend/node_modules/**",
      "frontend/dist/**",
    ],
  },

  ...vue.configs["flat/essential"],

  {
    files: ["frontend/**/*.{js,vue}"],

    plugins: {
      "@stylistic": stylistic,
    },

    rules: {
      "vue/multi-word-component-names": "off",

      "@stylistic/padding-line-between-statements": [
        "error",

        {
          blankLine: "never",
          prev: "*",
          next: "*",
        },

        {
          blankLine: "always",
          prev: "*",
          next: "function",
        },

        {
          blankLine: "always",
          prev: "function",
          next: "*",
        },
      ],
    },
  },
];
