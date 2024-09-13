package main

import "fmt"

type Person struct	{
	age	int
	name	string
	sex	string
}

func (person *Person) Introduce()	{
	fmt.Println("My name is ", person.name, " and My age is ", person.age, " also sex is ", person.sex)	
}

func (person *Person) Aging()	{
	person.age += 10
}

func main()	{
	person := Person{24, "SSH", "MAN"}

	person.Aging()

	person.Introduce()
}
