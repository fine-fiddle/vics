import unittest
import vics
import os

global test_db 
test_db = "testing_db.sqlite"


class TestVicsServer(unittest.TestCase):
    
    def test_create_new_database(self):
        vics.create_new_database(test_db)
        self.assertTrue(os.path.exists(test_db) == 1)
        os.remove(test_db)
