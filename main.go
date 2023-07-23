package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Task struct to match the fields in the JSON payload and database schema
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
}

func main() {
	r := gin.Default()

	createTable()

	// Define the routes
	r.POST("/tasks", createTaskHandler)
	r.GET("/tasks/:id", retrieveTaskHandler)
	r.PUT("/tasks/:id", updateTaskHandler)
	r.DELETE("/tasks/:id", deleteTaskHandler)
	r.GET("/tasks", listTasksHandler)

	r.Run(":8080")
}

// Create table
func createTable() {
	db, err := connectDB()
	if err != nil {
		log.Fatal("error connection to db")
	}
	defer db.Close()
	projects_table := `CREATE TABLE IF NOT EXISTS Project_Details (
			Id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
			Title TEXT NOT NULL,
			Description TEXT,
			Due_date TEXT,
			Status TEXT
			);`
	query, err := db.Prepare(projects_table)
	fmt.Println("Table created successfully!")
	if err != nil {
		log.Fatal(err)
		fmt.Println("Table already created!")
	}
	query.Exec()
}

// Database connection function
func connectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "my_task.db")
	fmt.Println("db connected successfully!")
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Part 1: Create a new task
func createTaskHandler(c *gin.Context) {
	var newTask Task
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		fmt.Println("Invalid JSON payload")
		return
	}

	// Validation on the required field
	if newTask.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}
	if newTask.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}
	if newTask.DueDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Due Date is required"})
		return
	}

	// Status validation for pending in progress and completed
	if newTask.Status != "Pending" || newTask.Status == "In Progress" || newTask.Status == "Completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status will be : Pending, In Progress, Completed."})
		return
	}

	// Validate the date format (DD-MM-YYYY)
	_, err := time.Parse("02/01/2006", newTask.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Due Date format. Use DD/MM/YYYY"})
		return
	}

	// Insert the new task into the database
	db, err := connectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO Project_Details (title, description, due_date, status) VALUES (?, ?, ?, ?)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare SQL statement"})
		fmt.Println("Failed to prepare SQL statement")
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(newTask.Title, newTask.Description, newTask.DueDate, newTask.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert task into the database"})
		fmt.Println("Failed to insert task into the database")
		return
	}

	taskID, _ := result.LastInsertId()
	newTask.ID = int(taskID)

	c.JSON(http.StatusCreated, newTask)
	fmt.Println("Project details inserted successfully!", newTask)
}

// Part 2: Retrieve a task
func retrieveTaskHandler(c *gin.Context) {
	taskID := c.Param("id")

	// Retrieve the task from the database
	db, err := connectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}
	defer db.Close()

	var task Task
	err = db.QueryRow("SELECT * FROM Project_Details WHERE id = ?", taskID).
		Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
	fmt.Println("Project get by ID", taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// Part 3: Update a task
func updateTaskHandler(c *gin.Context) {
	// Get the task ID from the request URL parameter
	taskIDStr := c.Param("id")

	// Check if the taskID is valid and parse it to an integer
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Retrieve the task from the database
	db, err := connectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}
	defer db.Close()

	// Query the task from the database using the ID
	row := db.QueryRow("SELECT * FROM Project_Details WHERE ID = ?", taskID)
	fmt.Println("----------++++---------", row)
	var task Task
	err = row.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
	fmt.Println("+++++++++++", task)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task"})
		return
	}

	// Parse the request JSON payload to get the updated task details
	var updatedTask Task
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Update the task details in the database
	stmt, err := db.Prepare("UPDATE Project_Details SET Title=?, Description=?, Due_Date=?, Status=? WHERE ID=?")
	fmt.Println("Project details updated", taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare SQL statement"})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(updatedTask.Title, updatedTask.Description, updatedTask.DueDate, updatedTask.Status, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Set the ID of the updated task to the original task ID
	updatedTask.ID = taskID

	c.JSON(http.StatusOK, updatedTask)
}

// // Part 3: Update a task
// func updateTaskHandler(c *gin.Context) {
// 	taskIDStr := c.Param("id")
// 	// Check if the taskID is valid and parse it to an integer
// 	taskID, err := strconv.Atoi(taskIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
// 		return
// 	}

// 	// Check if the task with the given ID exists in the database
// 	db, err := connectDB()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
// 		return
// 	}
// 	defer db.Close()

// 	// var updatedTask Task
// 	// if err := c.ShouldBindJSON(&updatedTask); err != nil {
// 	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
// 	// 	return
// 	// }

// 	// Query the task from the database using the ID
// 	row := db.QueryRow("SELECT * FROM tasks WHERE ID = ?", taskID)

// 	var task Task
// 	err = row.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
// 			return
// 		}
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task"})
// 		return
// 	}

// 	// // Update the task in the database
// 	// db, err := connectDB()
// 	// if err != nil {
// 	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
// 	// 	return
// 	// }
// 	// defer db.Close()

// 	// stmt, err := db.Prepare("UPDATE Project_Details SET title=?, description=?, due_date=?, status=? WHERE id=?")
// 	// if err != nil {
// 	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare SQL statement"})
// 	// 	return
// 	// }
// 	// defer stmt.Close()

// 	// _, err = stmt.Exec(updatedTask.Title, updatedTask.Description, updatedTask.DueDate, updatedTask.Status, taskID)
// 	// if err != nil {
// 	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task in the database"})
// 	// 	return
// 	// }

// 	// Parse the request JSON payload to get the updated task details
// 	var updatedTask Task
// 	if err := c.ShouldBindJSON(&updatedTask); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
// 		return
// 	}

// 	// Update the task details in the database
// 	stmt, err := db.Prepare("UPDATE tasks SET Title=?, Description=?, DueDate=?, Status=? WHERE ID=?")
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare SQL statement"})
// 		return
// 	}
// 	defer stmt.Close()

// 	_, err = stmt.Exec(updatedTask.Title, updatedTask.Description, updatedTask.DueDate, updatedTask.Status, taskID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
// 		return
// 	}

// 	// Set the ID of the updated task to the original task ID
// 	updatedTask.ID = taskID

// 	c.JSON(http.StatusOK, updatedTask)
// }

// Part 4: Delete a task
func deleteTaskHandler(c *gin.Context) {
	taskID := c.Param("id")

	// Delete the task from the database
	db, err := connectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}
	defer db.Close()

	// Query the task from the database using the ID
	row := db.QueryRow("SELECT * FROM Project_Details WHERE ID = ?", taskID)
	fmt.Println("Project not exist")
	var task Task
	err = row.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
	fmt.Println("+++++++++++", task)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task"})
		return
	}

	stmt, err := db.Prepare("DELETE FROM Project_Details WHERE id=?")
	fmt.Println("Project deleted successfully!")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare SQL statement"})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task from the database"})
		fmt.Println()
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// Part 5: List all Project_Details
func listTasksHandler(c *gin.Context) {
	// Retrieve all tasks from the database
	db, err := connectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, description, due_date, status FROM Project_Details")
	fmt.Println("Project List Shown")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Project_Details from the database"})
		return
	}
	defer rows.Close()

	var Project_Details []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan task from the database"})
			return
		}
		Project_Details = append(Project_Details, task)
	}

	c.JSON(http.StatusOK, Project_Details)
}
