package calculation

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

type Formula struct {
	Expression string
}

// Функция для вычисления математического выражения
func CalculateFormula(formula *Formula) (int, error) {
	// Удаляем все пробелы из выражения
	expression := formula.Expression
	expression = strings.ReplaceAll(expression, " ", "")

	// Проверяем входную строку на наличие недопустимых символов
	match, _ := regexp.MatchString(`^[0-9d()+\-*/\s]+$`, expression)
	if !match {
		return 0, errors.New("недопустимые символы в формуле")
	}

	// Заменяем символы d на случайные числа
	expression, err := replaceDiceRolls(expression)
	if err != nil {
		return 0, err
	}

	// Создаем стек для хранения операндов
	stack := []float64{}

	// Создаем стек для хранения операторов
	operators := []string{}

	// Функция для проверки приоритета операторов
	precedence := func(operator string) int {
		if operator == "+" || operator == "-" {
			return 1
		} else if operator == "*" || operator == "/" {
			return 2
		}
		return 0
	}

	// Функция для выполнения операции с двумя операндами и оператором
	performOperation := func() {
		operator := operators[len(operators)-1]
		operators = operators[:len(operators)-1]

		operand2 := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		operand1 := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		var result float64

		switch operator {
		case "+":
			result = operand1 + operand2
		case "-":
			result = operand1 - operand2
		case "*":
			result = operand1 * operand2
		case "/":
			result = operand1 / operand2
		}

		stack = append(stack, result)
	}

	// Итерируемся по каждому символу в выражении
	for i := 0; i < len(expression); i++ {
		ch := string(expression[i])

		// Если текущий символ - цифра, продолжаем считывать число
		if ch >= "0" && ch <= "9" {
			numberStr := ""
			for i < len(expression) && ((expression[i] >= '0' && expression[i] <= '9') || expression[i] == '.') {
				numberStr += string(expression[i])
				i++
			}
			i--
			number, err := strconv.ParseFloat(numberStr, 64)
			if err != nil {
				return 0, err
			}
			stack = append(stack, number)
		} else if ch == "(" {
			operators = append(operators, ch)
		} else if ch == ")" {
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				performOperation()
			}
			operators = operators[:len(operators)-1] // Удаляем "(" из стека операторов
		} else {
			// Если текущий символ - оператор
			for len(operators) > 0 && precedence(operators[len(operators)-1]) >= precedence(ch) {
				performOperation()
			}
			operators = append(operators, ch)
		}
	}

	// Выполняем оставшиеся операции в стеке операторов
	for len(operators) > 0 {
		performOperation()
	}

	// Результат будет находиться в верхушке стека операндов
	result := stack[len(stack)-1]

	return int(math.Floor(float64(result))), nil
}

// Generate random dice results and replace it into formula
func replaceDiceRolls(expression string) (string, error) {
	//rand.Seed(time.Now().UnixNano())

	//Check bad string after d character
	re := regexp.MustCompile(`d([^0-9]|$)`)
	matches := re.FindAllString(expression, -1)
	if matches != nil {
		return "", fmt.Errorf("недопустимая запись после символа d в выражении")
	}

	// Регулярное выражение для поиска выражений с символами d
	re = regexp.MustCompile(`(\d+)?d(\d+)`)

	// Заменяем выражения с символами d на случайные числа
	result := re.ReplaceAllStringFunc(expression, func(match string) string {
		diceParts := re.FindStringSubmatch(match)
		count, _ := strconv.Atoi(diceParts[1])
		if count == 0 {
			count = 1
		}
		sides, _ := strconv.Atoi(diceParts[2])

		// Генерируем случайные числа
		sum := 0
		for i := 0; i < count; i++ {
			roll := rand.Intn(sides) + 1
			sum += roll
		}

		return strconv.Itoa(sum)
	})

	return result, nil
}
