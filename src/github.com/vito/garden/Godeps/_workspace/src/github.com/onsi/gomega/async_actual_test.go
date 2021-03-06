package gomega

import (
	. "github.com/onsi/ginkgo"
	"time"
)

func init() {
	Describe("AsyncActual", func() {
		var (
			failureMessage string
			callerSkip     int
		)

		var fakeFailHandler = func(message string, skip ...int) {
			failureMessage = message
			callerSkip = skip[0]
		}

		BeforeEach(func() {
			failureMessage = ""
			callerSkip = 0
		})

		Describe("Eventually", func() {
			Context("when passed a function", func() {
				Context("the positive case", func() {
					It("should poll the function and matcher", func() {
						arr := []int{}
						a := newAsyncActual(asyncActualTypeEventually, func() []int {
							arr = append(arr, 1)
							return arr
						}, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.Should(HaveLen(10))

						Ω(arr).Should(HaveLen(10))
						Ω(failureMessage).Should(BeZero())
					})

					It("should continue when the matcher errors", func() {
						var arr = []int{}
						a := newAsyncActual(asyncActualTypeEventually, func() interface{} {
							arr = append(arr, 1)
							if len(arr) == 4 {
								return 0 //this should cause the matcher to error
							}
							return arr
						}, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.Should(HaveLen(4), "My description %d", 2)

						Ω(failureMessage).Should(ContainSubstring("Timed out after"))
						Ω(failureMessage).Should(ContainSubstring("My description 2"))
						Ω(callerSkip).Should(Equal(4))
					})

					It("should be able to timeout", func() {
						arr := []int{}
						a := newAsyncActual(asyncActualTypeEventually, func() []int {
							arr = append(arr, 1)
							return arr
						}, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.Should(HaveLen(11), "My description %d", 2)

						Ω(arr).Should(HaveLen(10))
						Ω(failureMessage).Should(ContainSubstring("Timed out after"))
						Ω(failureMessage).Should(ContainSubstring("My description 2"))
						Ω(callerSkip).Should(Equal(4))
					})
				})

				Context("the negative case", func() {
					It("should poll the function and matcher", func() {
						counter := 0
						arr := []int{}
						a := newAsyncActual(asyncActualTypeEventually, func() []int {
							counter += 1
							if counter >= 10 {
								arr = append(arr, 1)
							}
							return arr
						}, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.ShouldNot(HaveLen(0))

						Ω(arr).Should(HaveLen(1))
						Ω(failureMessage).Should(BeZero())
					})

					It("should timeout when the matcher errors", func() {
						a := newAsyncActual(asyncActualTypeEventually, func() interface{} {
							return 0 //this should cause the matcher to error
						}, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.ShouldNot(HaveLen(0), "My description %d", 2)

						Ω(failureMessage).Should(ContainSubstring("Timed out after"))
						Ω(failureMessage).Should(ContainSubstring("Error:"))
						Ω(failureMessage).Should(ContainSubstring("My description 2"))
						Ω(callerSkip).Should(Equal(4))
					})

					It("should be able to timeout", func() {
						a := newAsyncActual(asyncActualTypeEventually, func() []int {
							return []int{}
						}, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.ShouldNot(HaveLen(0), "My description %d", 2)

						Ω(failureMessage).Should(ContainSubstring("Timed out after"))
						Ω(failureMessage).Should(ContainSubstring("My description 2"))
						Ω(callerSkip).Should(Equal(4))
					})
				})
			})
		})

		Describe("Consistently", func() {
			Describe("The positive case", func() {
				Context("when the matcher consistently passes for the duration", func() {
					It("should pass", func() {
						calls := 0
						a := newAsyncActual(asyncActualTypeConsistently, func() string {
							calls++
							return "foo"
						}, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.Should(Equal("foo"))
						Ω(calls).Should(Equal(10))
						Ω(failureMessage).Should(BeZero())
					})
				})

				Context("when the matcher fails at some point", func() {
					It("should fail", func() {
						calls := 0
						a := newAsyncActual(asyncActualTypeConsistently, func() interface{} {
							calls++
							if calls > 9 {
								return "bar"
							}
							return "foo"
						}, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.Should(Equal("foo"))
						Ω(failureMessage).Should(ContainSubstring("to equal"))
						Ω(callerSkip).Should(Equal(4))
					})
				})

				Context("when the matcher errors at some point", func() {
					It("should fail", func() {
						calls := 0
						a := newAsyncActual(asyncActualTypeConsistently, func() interface{} {
							calls++
							if calls > 5 {
								return 3
							}
							return []int{1, 2, 3}
						}, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.Should(HaveLen(3))
						Ω(failureMessage).Should(ContainSubstring("HaveLen matcher expects"))
						Ω(callerSkip).Should(Equal(4))
					})
				})
			})

			Describe("The negative case", func() {
				Context("when the matcher consistently passes for the duration", func() {
					It("should pass", func() {
						c := make(chan bool)
						a := newAsyncActual(asyncActualTypeConsistently, c, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.ShouldNot(Receive())
						Ω(failureMessage).Should(BeZero())
					})
				})

				Context("when the matcher fails at some point", func() {
					It("should fail", func() {
						c := make(chan bool)
						go func() {
							time.Sleep(time.Duration(100 * time.Millisecond))
							c <- true
						}()

						a := newAsyncActual(asyncActualTypeConsistently, c, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.ShouldNot(Receive())
						Ω(failureMessage).Should(ContainSubstring("not to receive anything"))
					})
				})

				Context("when the matcher errors at some point", func() {
					It("should fail", func() {
						calls := 0
						a := newAsyncActual(asyncActualTypeConsistently, func() interface{} {
							calls++
							return calls
						}, fakeFailHandler, time.Duration(0.2*float64(time.Second)), time.Duration(0.02*float64(time.Second)), 1)

						a.ShouldNot(BeNumerically(">", 5))
						Ω(failureMessage).Should(ContainSubstring("not to be >"))
						Ω(callerSkip).Should(Equal(4))
					})
				})
			})
		})

		Context("when passed a function with the wrong # or arguments & returns", func() {
			It("should panic", func() {
				Ω(func() {
					newAsyncActual(asyncActualTypeEventually, func() {}, fakeFailHandler, 0, 0, 1)
				}).Should(Panic())

				Ω(func() {
					newAsyncActual(asyncActualTypeEventually, func(a string) int { return 0 }, fakeFailHandler, 0, 0, 1)
				}).Should(Panic())

				Ω(func() {
					newAsyncActual(asyncActualTypeEventually, func() int { return 0 }, fakeFailHandler, 0, 0, 1)
				}).ShouldNot(Panic())
			})
		})
	})
}
