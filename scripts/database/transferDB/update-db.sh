#!/bin/bash

# Step 1: Create new tables
echo "Step 1: Creating new tables..."
docker exec -i triggerx-scylla cqlsh -e "USE triggerx; DESCRIBE TABLES;" || {
    echo "Failed to connect to database"
    exit 1
}

# Create new tables
docker exec -i triggerx-scylla cqlsh < scripts/database/transferDB/update-schema-step1.cql
if [ $? -eq 0 ]; then
    echo "New tables created successfully"
    
    # Step 2: Transfer data using Python script
    echo "Step 2: Transferring data..."
    python3 scripts/database/transferDB/transfer_data.py
    
    if [ $? -eq 0 ]; then
        echo "Data transfer completed successfully"
        
        # Verify data transfer
        echo "Verifying data transfer..."
        docker exec -i triggerx-scylla cqlsh -e "USE triggerx; SELECT COUNT(*) FROM keeper_data; SELECT COUNT(*) FROM keeper_data_new; SELECT COUNT(*) FROM user_data; SELECT COUNT(*) FROM user_data_new;"
        
        read -p "Does the data look correct? (yes/no) " answer
        if [ "$answer" != "yes" ]; then
            echo "Aborting update process"
            exit 1
        fi
        
        # Step 3: Create new tables with original names
        echo "Step 3: Creating new tables with original names..."
        docker exec -i triggerx-scylla cqlsh < scripts/database/transferDB/update-schema-step2.cql
        
        if [ $? -eq 0 ]; then
            echo "New tables created successfully"
            
            # Step 4: Transfer data from new tables to original tables
            echo "Step 4: Transferring data to original tables..."
            python3 scripts/database/transferDB/transfer_data_step2.py
            
            if [ $? -eq 0 ]; then
                echo "Data transfer completed successfully"
                
                # Verify final data
                echo "Verifying final data..."
                docker exec -i triggerx-scylla cqlsh -e "USE triggerx; SELECT COUNT(*) FROM keeper_data; SELECT COUNT(*) FROM user_data;"
                
                read -p "Does the final data look correct? (yes/no) " answer
                if [ "$answer" != "yes" ]; then
                    echo "Aborting update process"
                    exit 1
                fi
                
                # Step 5: Drop temporary tables
                echo "Step 5: Dropping temporary tables..."
                docker exec -i triggerx-scylla cqlsh < scripts/database/transferDB/update-schema-step3.cql
                
                if [ $? -eq 0 ]; then
                    echo "Database schema updated successfully"
                    echo "Verifying final tables..."
                    docker exec -i triggerx-scylla cqlsh -e "USE triggerx; DESCRIBE TABLES;"
                else
                    echo "Failed to drop temporary tables"
                    exit 1
                fi
            else
                echo "Failed to transfer data to original tables"
                exit 1
            fi
        else
            echo "Failed to create new tables with original names"
            exit 1
        fi
    else
        echo "Failed to transfer data"
        exit 1
    fi
else
    echo "Failed to create new tables"
    exit 1
fi