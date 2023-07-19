package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"log"

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

func createTable() {
	db, err := connectDB()
	if err != nil {
		log.Fatal("error conneting to db")
	}
	defer db.Close()
	tasks_table := `CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            description TEXT,
            due_date DATE,
            status TEXT
			);`
	query, err := db.Prepare(tasks_table)
	if err != nil {
		log.Fatal(err)
	}
	query.Exec()
	fmt.Println("Table created successfully!")
	return
}

// Database connection function
func connectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "my_assignment.db")
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
		return
	}

	// Insert the new task into the database
	db, err := connectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO tasks (title, description, due_date, status) VALUES (?, ?, ?, ?)")
	fmt.Println(stmt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare SQL statement"})
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(newTask.Title, newTask.Description, newTask.DueDate, newTask.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert task into the database"})
		return
	}

	taskID, _ := result.LastInsertId()
	newTask.ID = int(taskID)

	c.JSON(http.StatusCreated, newTask)
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
	err = db.QueryRow("SELECT id, title, description, due_date, status FROM tasks WHERE id = ?", taskID).
		Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// Part 3: Update a task
func updateTaskHandler(c *gin.Context) {
	taskID := c.Param("id")

	var updatedTask Task
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Update the task in the database
	db, err := connectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE tasks SET title=?, description=?, due_date=?, status=? WHERE id=?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare SQL statement"})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(updatedTask.Title, updatedTask.Description, updatedTask.DueDate, updatedTask.Status, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task in the database"})
		return
	}

	c.JSON(http.StatusOK, updatedTask)
}

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

	stmt, err := db.Prepare("DELETE FROM tasks WHERE id=?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare SQL statement"})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task from the database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// Part 5: List all tasks
func listTasksHandler(c *gin.Context) {
	// Retrieve all tasks from the database
	db, err := connectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, description, due_date, status FROM tasks")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks from the database"})
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan task from the database"})
			return
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, tasks)
}
