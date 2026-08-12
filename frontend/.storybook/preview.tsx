import type { Preview } from "@storybook/react-vite";

import "../src/styles.css";

const preview: Preview = {
  decorators: [
    (Story) => {
      document.documentElement.lang = "ja";
      return <Story />;
    },
  ],
  parameters: {
    layout: "padded",
    a11y: { test: "error" },
  },
  tags: ["autodocs"],
};

export default preview;
