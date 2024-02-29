class LargeObject:
    def __init__(self):
        # Create a large string
        self.data = '0' * (10 ** 8)  

# Create a list of large objects
large_objects = [LargeObject() for _ in range(3)]
