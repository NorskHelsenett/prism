// Create a utility function for number formatting
function formatNumber(value) {
  // Create a NumberFormat object for US English, which uses comma separators
  const formatter = new Intl.NumberFormat('en-US', {
    maximumFractionDigits: 2, // Adjust this for desired number of decimal places
  });

  return formatter.format(value);
}

export default formatNumber;