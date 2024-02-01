export function textAreaListHelper(node) {
  // Define the helper function
  function handleTextareaInput(event) {
    if (event.key === 'Enter') {
      const textarea = event.target;
      const cursorPosition = textarea.selectionStart;
      const beforeCursor = textarea.value.substring(0, cursorPosition);
      const afterCursor = textarea.value.substring(cursorPosition);
      const lines = beforeCursor.split('\n');
      const lastLine = lines[lines.length - 1];

      // Define the patterns to check in an array, ordered from longest to shortest
      const patterns = ['- [ ] ', '- '];

      // Loop through patterns to find a match
      for (const pattern of patterns) {
        const patternLength = pattern.length;
        if (lastLine.trim() === pattern.trim()) {
          // Check if the content after the cursor starts with a newline
          const newLineToAdd = afterCursor.startsWith('\n') ? '' : '\n';
          textarea.value = lines.slice(0, -1).join('\n') + newLineToAdd + afterCursor;
          // Adjust cursor position to be at the start of the new line
          textarea.selectionStart = textarea.selectionEnd = beforeCursor.length - lastLine.length + newLineToAdd.length;
          // event.preventDefault();
          break;
        } else if (lastLine.startsWith(pattern.trim())) {
          event.preventDefault();
          textarea.value = beforeCursor + '\n' + pattern + afterCursor;
          const newPosition = beforeCursor.length + patternLength + 1;
          textarea.selectionStart = textarea.selectionEnd = newPosition;
          break;
        }
      }
    }
  }

  // Attach the event listener
  node.addEventListener('keydown', handleTextareaInput);

  // Return a destroy function to clean up
  return {
    destroy() {
      node.removeEventListener('keydown', handleTextareaInput);
    }
  };
}
