package Test

//func main() {
//	env := replier.NewEnv()
//	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", env.RabbitmqUsername, env.RabbitmqPassword, env.RabbitmqUrl))
//	go replier.Reply()
//
//	submission := replier.SubmissionMessage{
//		SubmissionID:   "123",
//		ProblemID:      "14",
//		UserID:         "123",
//		TestCaseNumber: 10,
//		Prompt:         "i need a Python function that includes the following logic: The base form of the function for 0 input is 5.If the input 'i' is even, the function should be func(i−1)−21. If 'i' is odd, the function should be func(i−1)**2. Additionally, the function should print the final output at the end. and also get the input from the user and don't write any example and get the input without any text and print the output without any text and 10000 limit recursion",
//	}
//
//	for i := 0; i < 5; i++ {
//		go func() {
//			err := request(submission)
//			if err != nil {
//				return
//			}
//		}()
//	}
//	ch, err := conn.Channel()
//	if err != nil {
//		return
//	}
//	msgs, err := ch.Consume("result", "", false, false, false, false, nil)
//	start := time.Now()
//	check := make(map[string]bool)
//	fmt.Println("Waiting for results")
//	for msg := range msgs {
//		msg := msg
//		go func() {
//			fmt.Println(string(msg.Body))
//			value := make(map[string]interface{})
//			json.Unmarshal(msg.Body, &value)
//			fmt.Println(value["results"])
//			check[value["user_id"].(string)] = true
//			msg.Ack(true)
//		}()
//		if len(check) == 5 {
//			since := time.Since(start)
//			fmt.Println(since)
//			fmt.Println(check)
//			break
//		}
//	}
//	select {}
//}

//func Request(submission replier.SubmissionMessage) {
//	url := "https://fastapi-kodalab.chbk.run/api/v1/prompt/"
//	requestBody, err := json.Marshal(submission)
//	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
//	if err != nil {
//		fmt.Println("Error creating request:", err)
//		return
//	}
//
//	// Set the Content-Type header
//	req.Header.Set("Content-Type", "application/json")
//
//	// Create a new HTTP client
//	client := &http.Client{}
//
//	// Send the request
//	resp, err := client.Do(req)
//	if err != nil {
//		fmt.Println("Error sending request:", err)
//		return
//	}
//	defer resp.Body.Close()
//
//	// Check the response status code
//	if resp.StatusCode != http.StatusOK {
//		fmt.Println("Unexpected status code:", resp.StatusCode)
//		return
//	}
//	fmt.Println("Submission sent")
//}
//
//func main() {
//
//	submission := replier.SubmissionMessage{
//		SubmissionID:   "123",
//		ProblemID:      "14",
//		UserID:         "123",
//		TestCaseNumber: 10,
//		Prompt:         "i need a Python function that includes the following logic: The base form of the function for 0 input is 5.If the input 'i' is even, the function should be func(i−1)−21. If 'i' is odd, the function should be func(i−1)**2. Additionally, the function should print the final output at the end. and also get the input from the user and don't write any example and get the input without any text and print the output without any text and 10000 limit recursion",
//	}
//
//	for i := 0; i < 5; i++ {
//		submission.UserID = fmt.Sprintf("%d", i)
//		Request(submission)
//	}
//	select {}
//}
