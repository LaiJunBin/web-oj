// mock complex javascript library code for sonar analysis

// function add(a, b) {return a + b;}

// function subtract(a, b) {return a - b;}

// function multiply(a, b) {return a * b;}

// function divide(a, b) {return a / b;}

// function modulus(a, b) {return a % b;}

// function power(a, b) {return Math.pow(a, b);}

// function squareRoot(a) {return Math.sqrt(a);}

// function log(a) {return Math.log(a);}

// function exp(a) {return Math.exp(a);}

// function sin(a) {return Math.sin(a);}

// function cos(a) {return Math.cos(a);}


// make above code to uglify and single line

const add = (a, b) => a + b, 
subtract = (a, b) => a - b, 
multiply = (a, b) => a * b, 
divide = (a, b) => a / b, 
modulus = (a, b) => a % b, 
power = (a, b) => Math.pow(a, b), 
squareRoot = (a) => Math.sqrt(a);

// mock complex calling to bad code smell

const result = add(1, 2) + subtract(3, 4) * multiply(5, 6) / divide(7, 8) % modulus(9, 10) ** power(11, 12) + squareRoot(13);

console.log(result);
