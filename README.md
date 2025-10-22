# Safe Intelligence - Full Stack Engineer - Code Assessment

This assessment is designed to evaluate your practical skills and understanding of full-stack development, with a focus on TypeScript/React for the frontend, Golang for the backend, and API-first development using OpenAPI. We're looking for clean, well-tested, and maintainable code, along with an understanding of modern software development practices.

You will be given 2 days to complete this assessment.

## 🚀 The Challenge: Todo Application

Your task is to build a basic full-stack Todo application. To keep the focus on core development skills and reduce setup complexity, the backend will use an in-memory data store for the todo items.

What has been provided:
- An OpenAPI specification file (`openapi.yaml`) that defines a health check endpoint.
- A Golang backend with a basic health check endpoint.
- A React frontend that uses the Mantine component library and only a basic header component with a health check display.

## ✨ Requirements

Please implement the following features and ensure your solution meets the specified criteria:

* Extend the provided OpenAPI specification to include endpoints for GET and POST operations on todo items at minimum.
* Implement the Golang backend to handle these endpoints, this should allow you to create Todo items and retrieve a list of all Todo items at minimum.
* Extend the React frontend to display a list of todo items and allow users to add new todo items.
* Dockerise both the Golang backend and the React frontend.

These are the basic requirements, you can extend the functionality as you see fit, but please ensure that the core requirements are met first.

## 🖥️ System Requirements

* **Golang**: Ensure you have Go installed (version 1.24.2 or later).
* **Node.js**: Ensure you have Node.js installed (version 23 or later).

## 🚀 Usage

To generate the backend and frontend code from the OpenAPI specification, you can use the following commands:

```bash
# For the Golang backend
go generate ./...

# For the React frontend
npm run openapi
```

To run the backend you can use:

```bash
go run cmd/main.go
```

To run the frontend, you can use:

```bash
cd frontend

# Make sure to install dependencies first
npm install 

# Then start the development server
npm run dev
```

The frontend will be available at `http://localhost:5173` and the backend at `http://localhost:8080`. With both the backend and frontend running, you should be able to access the application in your web browser and see this page:

![Todo Application Screenshot](https://github.com/user-attachments/assets/82ed8f26-3acb-492d-bc4a-feebb02e0cbc)

## 💡 Key Considerations

While there is a lot of flexibility in how you implement the solution, we are particularly interested in how clean, maintainable, and testable your code is. Here are some key considerations to keep in mind:

* **Code Quality**: Write clean, readable, and well-structured code. Adhere to conventional style guides for both Go and TypeScript.
* **Error Handling**: Implement robust error handling on both the frontend and backend, providing meaningful error messages and appropriate HTTP status codes.
* **Testing**: Demonstrate a strong commitment to code quality through comprehensive unit and integration tests.

## 📦 Submission Instructions

* When you have completed the assessment, please ensure your code is well-organized and follows best practices.
* Please document your code in the REVIEW.md file, including:
  * Brief detail of any changes you made.
  * Your approach to solving the problem.
  * Any additional improvements you would make if given more time.
* Create a ZIP archive of your solution, including the frontend and backend code, Dockerfiles, and any other relevant files.
* Email the ZIP archive back to us.

Good luck with the assessment! We look forward to seeing your solution and learning about your approach to full-stack development. If you have any questions or need clarification on any aspect of the assessment, please don't hesitate to reach out.
