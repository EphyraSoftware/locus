import daisyui from "daisyui"

/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{vue,ts}",
    ],
    theme: {
        extend: {},
    },
    daisyui: {
        themes: ["synthwave"]
    },
    plugins: [
        daisyui,
    ],
}
