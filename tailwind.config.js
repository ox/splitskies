module.exports = {
  content: ["./templates/*.{html,js}", "./templates/**/*.{html,js}"],
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ]
}