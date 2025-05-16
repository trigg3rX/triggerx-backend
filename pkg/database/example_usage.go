package database

// // Example of how to use separate connections for each ScyllaDB node
// func ExampleNodeConnections() {
// 	// Connection to ScyllaDB node 1
// 	scylla1Config := NewScylla1Config()
// 	scylla1Conn, err := NewConnection(scylla1Config)
// 	if err != nil {
// 		// Handle error
// 		return
// 	}
// 	defer scylla1Conn.Close()

// 	// Connection to ScyllaDB node 2
// 	scylla2Config := NewScylla2Config()
// 	scylla2Conn, err := NewConnection(scylla2Config)
// 	if err != nil {
// 		// Handle error
// 		return
// 	}
// 	defer scylla2Conn.Close()

// 	// Example: Use specific node connections for different operations
// 	// For read operations on node 1
// 	scylla1Session := scylla1Conn.Session()
// 	_ = scylla1Session // Used for read operations

// 	// For write operations on node 2
// 	scylla2Session := scylla2Conn.Session()
// 	_ = scylla2Session // Used for write operations
// }
