package db

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"math/rand"

	"github.com/addvanced/gophersocial/internal/store"
)

var (
	usernames = []string{
		"JohnDoe", "Kenneth", "Alice", "Bob", "Charlie", "David", "Eve",
		"Frank", "Grace", "Heidi", "Ivan", "Judy", "Mallory", "Oscar",
		"Peggy", "Romeo", "Trent", "Victor", "Walter", "Zoe",
		"Ada", "Bjarne", "Charles", "Dennis", "Elon", "Grace",
		"Haskell", "Isaac", "Julia", "Ken", "Linus", "Mary",
		"Niklaus", "Ole-Johan", "Pascal", "Quentin", "Rasmus", "Simon",
		"Tim", "Ursula", "Vint", "Woz", "Xavier", "Yukihiro",
	}

	titles = []string{
		"10 Tips to Improve Your Productivity",
		"The Ultimate Guide to Remote Work",
		"How to Stay Focused in a Distracting World",
		"Mastering Time Management: A Complete Guide",
		"Top 5 Tools for Better Project Management",
		"The Future of Artificial Intelligence in Business",
		"Understanding Blockchain Technology",
		"Why Cybersecurity is Crucial in 2024",
		"How to Create a Successful Online Business",
		"The Power of Networking in the Digital Age",
		"Building a Personal Brand: A Step-by-Step Guide",
		"Effective Communication Strategies for Remote Teams",
		"How to Start a Podcast: A Beginner's Guide",
		"The Importance of Mental Health in the Workplace",
		"Top Marketing Strategies for Small Businesses",
		"Design Thinking: Solving Problems Creatively",
		"The Best Practices for Building Scalable Web Apps",
		"How to Use Social Media to Grow Your Business",
		"The Role of Data Analytics in Business Decision Making",
		"Why Emotional Intelligence is Key to Leadership",
		"Top 10 Trends in the Tech Industry for 2024",
		"The Evolution of E-Commerce: What You Need to Know",
		"How to Develop a Growth Mindset",
		"Digital Transformation: A Comprehensive Guide",
		"Why UX Design is Critical to Product Success",
		"Effective Strategies for Customer Retention",
		"Building a Strong Company Culture in a Remote World",
		"The Rise of Subscription-Based Business Models",
		"How to Manage Finances as a Freelancer",
		"Cybersecurity Best Practices for Small Businesses",
		"How to Build a Strong Professional Network",
		"The Future of Work: What to Expect in 2025",
		"Introduction to Cloud Computing for Beginners",
		"Creating High-Performing Teams: A Leadership Guide",
		"How to Optimize Your Website for Search Engines",
		"The Benefits of Adopting a DevOps Culture",
		"Why Data Privacy Matters More Than Ever",
		"How to Use Automation to Streamline Business Processes",
		"The Importance of Emotional Intelligence in the Workplace",
		"Top Content Marketing Strategies for 2024",
		"How to Stay Ahead of the Curve in the Tech Industry",
		"Building a Sustainable Business Model",
		"The Art of Negotiation: Tips for Success",
		"Understanding the Basics of Machine Learning",
		"Why Design Thinking is Crucial for Innovation",
		"How to Develop Leadership Skills in the Tech Industry",
		"The Importance of Continuous Learning in Career Growth",
		"How to Balance Work and Life in the Modern World",
		"Effective Strategies for Product Launch Success",
	}

	contents = []string{
		"In today's fast-paced digital world, productivity is key to staying competitive. Whether you're working remotely or in an office, optimizing your time is crucial. Here are some essential tips to improve your focus, eliminate distractions, and get more done in less time. Start with setting clear goals and breaking down tasks into manageable chunks.",
		"Artificial intelligence is transforming industries at an unprecedented rate. From healthcare to finance, AI is revolutionizing the way businesses operate. In this article, we'll explore the latest AI trends and discuss how companies are leveraging this technology to gain a competitive edge. Learn about machine learning, neural networks, and more.",
		"Remote work is becoming the norm for many businesses worldwide. With this shift, it's important to establish effective communication and collaboration practices. In this post, we'll dive into the best tools for remote teams, how to maintain productivity, and the importance of work-life balance when working from home.",
		"Blockchain technology is not just about cryptocurrencies anymore. It's being used in various sectors such as healthcare, supply chain, and even real estate. This article explains the basics of blockchain, its applications, and why it's considered one of the most secure technologies available today.",
		"Cloud computing has changed the way businesses operate by offering scalable, on-demand computing resources. Whether you're a startup or a large corporation, leveraging cloud services can help you reduce costs and increase efficiency. Let's explore the different cloud service models and how they can benefit your business.",
		"Design thinking is a human-centered approach to problem-solving that encourages creativity and innovation. It's widely used in industries from tech to education. This article explores the core principles of design thinking and provides a step-by-step guide on how you can use it to create better products and services.",
		"E-commerce is booming, and more businesses are going online every day. But how do you stand out in a crowded market? In this post, we'll discuss strategies for growing your online business, from SEO optimization to creating a seamless user experience. Discover what it takes to succeed in the world of e-commerce.",
		"Leadership in the digital age requires a new set of skills. From emotional intelligence to adaptability, today's leaders need to be more versatile than ever. This article explores the traits that make effective leaders and how you can develop these skills to lead your team through change and uncertainty.",
		"Cybersecurity threats are evolving, and businesses must stay vigilant. This post covers the latest trends in cybersecurity and provides practical tips for protecting your company's data. From phishing scams to ransomware, learn how to safeguard your digital assets against malicious attacks.",
		"Content marketing is all about providing valuable information to your audience. But with so much content out there, how do you make yours stand out? This article explores content marketing strategies that work in 2024, from storytelling to data-driven content creation. Discover how to connect with your audience on a deeper level.",
		"Effective project management is the backbone of any successful project. Whether you're working on a small team or leading a large initiative, staying organized is key. This post provides a comprehensive guide to project management tools, methodologies, and best practices to ensure your project's success.",
		"Personal branding is more important than ever in today's digital world. Your online presence can make or break your career. In this post, we'll dive into strategies for building a personal brand that resonates with your target audience and sets you apart from the competition.",
		"Data privacy has become a top concern for businesses and consumers alike. With increasing regulations like GDPR, companies must take extra steps to ensure customer data is protected. In this article, we'll explore data privacy best practices and how your business can stay compliant with the latest laws.",
		"Scaling a business is no easy task. Whether you're expanding into new markets or growing your team, scaling requires careful planning and execution. This post provides practical tips for scaling your business without losing sight of what made you successful in the first place.",
		"Freelancing offers flexibility and freedom, but it also comes with its own set of challenges. In this article, we'll discuss the pros and cons of freelancing, tips for managing your time effectively, and how to maintain a steady stream of clients. Whether you're just starting out or a seasoned freelancer, this guide is for you.",
		"The gig economy is changing the way people work. More and more individuals are turning to freelance gigs for flexibility and control over their careers. In this post, we'll explore how the gig economy is reshaping industries and what you need to know if you're considering making the switch.",
		"Emotional intelligence (EQ) is a critical skill in both personal and professional life. This article explains why EQ matters more than ever and provides actionable tips for improving your emotional intelligence. Learn how to manage your emotions, build better relationships, and navigate challenging situations with ease.",
		"The world of mobile apps is constantly evolving. From new development frameworks to design trends, staying updated is crucial for app developers. This article covers the latest in mobile app development and provides tips on how to create high-quality, user-friendly apps.",
		"Creating a strong company culture is key to retaining talent and driving business success. This post dives into how to foster a positive work environment, from building trust to encouraging collaboration. Learn what it takes to create a culture that supports growth and innovation.",
		"Agile methodology has revolutionized the way teams approach project management. By breaking projects into smaller, manageable pieces, teams can work more efficiently and adapt to changes quickly. This article provides an overview of Agile principles and how to implement them in your organization for better results.",
	}

	tags = []string{
		"Technology", "Productivity", "Startups", "Leadership", "Marketing",
		"Finance", "Remote Work", "AI", "Machine Learning", "Cloud Computing",
		"Web Development", "Mobile Apps", "Entrepreneurship", "E-commerce",
		"Social Media", "Business Growth", "Cybersecurity", "Blockchain",
		"SEO", "Data Analytics", "UX Design", "Innovation", "Software Development",
		"Digital Transformation", "DevOps", "Customer Experience", "Branding",
		"Automation", "SaaS", "Freelancing", "Content Marketing", "Time Management",
		"Networking", "Collaboration", "Career Development", "Personal Growth",
		"Mindset", "Strategy", "Team Building", "Agile Methodology", "User Research",
		"Remote Collaboration", "Health & Wellness", "Work-Life Balance",
		"Creativity", "Problem Solving", "Growth Hacking", "Sustainability",
		"Project Management", "Artificial Intelligence", "Big Data", "Data Privacy",
		"Fintech", "Healthcare Tech", "Sales", "Customer Retention", "Design Thinking",
		"Productivity Hacks", "Coding", "Open Source", "Virtual Reality", "Augmented Reality",
		"Ethics", "Diversity", "Inclusion", "Personal Branding", "Online Courses",
		"Digital Marketing", "Creative Thinking", "Mobile Optimization", "Website Design",
		"User Engagement", "Team Leadership", "Self-improvement", "Psychology",
		"Emotional Intelligence", "Human Resources", "Recruitment", "Talent Management",
		"Online Learning", "Video Marketing", "Subscription Models", "Cloud Security",
		"Privacy Laws", "Customer Success", "Social Responsibility", "Green Tech",
		"Impact Investing", "Venture Capital", "Business Analytics", "Scaling Businesses",
		"User Acquisition", "Innovation Strategy", "Risk Management", "Mobile First",
		"Digital Identity", "Data Governance", "B2B Marketing", "Digital Tools",
		"Ethical Hacking", "Workplace Culture", "Future of Work", "Legal Tech",
		"Customer Loyalty", "Tech Trends", "Financial Planning", "SaaS Tools",
	}

	comments = []string{
		"Great article! Really enjoyed the insights.",
		"Thanks for sharing this, very helpful!",
		"Interesting perspective, I hadn't thought of it that way.",
		"Love the tips, will definitely try them out.",
		"Excellent post, keep up the good work!",
		"Very informative, learned a lot from this.",
		"This was exactly what I needed, thank you!",
		"Clear and concise, great job!",
		"Can't wait to apply these strategies.",
		"Awesome read! Looking forward to more posts like this.",
		"I appreciate the in-depth analysis.",
		"Solid advice, thanks for breaking it down.",
		"Helpful and to the point, much appreciated!",
		"This gave me some great ideas, thanks!",
		"Thanks for the clear explanation.",
		"Fantastic content, well done!",
		"Your writing style is engaging, I love it!",
		"Thanks for this, very practical advice.",
		"Interesting take, thanks for sharing.",
		"I've bookmarked this for future reference.",
		"This is so insightful, thank you!",
		"Excellent breakdown of a complex topic.",
		"Great tips, I'm going to try them out!",
		"I appreciate the clarity in your writing.",
		"Very helpful guide, thanks!",
		"This really resonated with me.",
		"Well written and easy to understand.",
		"Thanks for making this topic simple.",
		"This post was a game changer for me.",
		"Good points, I hadn't considered those before.",
		"Such a helpful article, thank you!",
		"Your tips are always so practical!",
		"I love how you explain things in a simple way.",
		"This helped me get unstuck, thanks!",
		"Awesome content as always!",
		"Looking forward to trying this out.",
		"Great resource, thanks for putting this together.",
		"This was a great read, thank you!",
		"Love the examples you used here.",
		"Really appreciate this detailed breakdown.",
		"This helped clear up a lot of confusion, thanks!",
		"I've shared this with my team, very useful.",
		"Your posts are always so informative!",
		"Thanks for the thoughtful insights.",
		"Appreciate the practical advice.",
		"Simple but effective tips, love it!",
		"Your content is always on point!",
		"Thanks for the awesome advice!",
		"This was so helpful, keep them coming!",
		"Great approach, I'm going to implement this.",
		"Really informative, thanks for sharing.",
		"This is a topic I've struggled with, thanks for shedding light on it.",
		"Your posts always provide so much value!",
		"Thanks for making this easy to understand.",
		"Such a useful post, will be referencing it again!",
		"I've been searching for something like this, thank you!",
		"Wonderful content, keep up the great work!",
		"This was an eye-opener, thanks!",
		"Thanks for providing such clear examples.",
		"I can't wait to share this with my colleagues.",
		"This helped me out a lot, thank you!",
		"Your writing is always so insightful.",
		"Perfect timing, I needed this today!",
		"Appreciate the detailed advice!",
		"Good stuff, thanks for sharing!",
		"Such a great resource, thank you!",
		"Really valuable content, thanks!",
		"I learned so much from this post.",
		"Thanks for taking the time to explain this.",
		"This is exactly what I was looking for, thanks!",
		"Love how you simplify complex topics.",
		"Great work, keep it up!",
		"I'm going to use these tips right away!",
		"This was very eye-opening, thank you!",
		"Such a practical guide, thanks!",
		"I always look forward to your posts.",
		"Really appreciate the actionable steps.",
		"Thanks for making this so clear!",
		"Great insights, thank you for sharing!",
		"This gave me a new perspective, thanks!",
		"I learned a lot from reading this, thanks!",
		"Very well explained, I really appreciate it.",
		"Love the practical approach here!",
		"I'll definitely be implementing this advice.",
		"Such valuable advice, thanks!",
		"Great content, as usual!",
		"Your posts are always so well-researched.",
		"Thank you for making this so understandable.",
		"This has really helped me out, thanks!",
		"I always get value from your articles.",
		"Thanks for the step-by-step approach.",
		"Solid post, thanks for the insight!",
		"I appreciate the clarity and depth here.",
		"Your content is consistently top-notch.",
		"This has been incredibly helpful, thanks!",
		"Fantastic advice, I'll definitely try this out.",
		"I learned a lot, thanks for sharing!",
		"Really clear and practical, love it!",
		"Excellent, I'm looking forward to more content from you.",
		"This article really spoke to me, thank you!",
	}
)

func Seed(ctx context.Context, store store.Storage) {
	var (
		userNum   = 200
		followNum = 10000

		postNum    = 200
		commentNum = 500000
	)

	store.Logger.Infoln("Seeding database...")

	store.Logger.Infof(" - Generating %d users...\n", userNum)
	users := generateUsers(userNum)
	if err := store.Users.CreateBatch(ctx, users); err != nil {
		store.Logger.Errorw("Could not create users", "error", err.Error())
		return
	}

	store.Logger.Infof(" - Generating %d followers...\n", followNum)
	follows := generateFollowers(followNum, users)
	if err := store.Follow.CreateBatch(ctx, follows); err != nil {
		store.Logger.Errorw("Could not create follows", "error", err.Error())
		return
	}

	store.Logger.Infof(" - Generating %d posts...\n", postNum)
	posts := generatePosts(postNum, users)
	if err := store.Posts.CreateBatch(ctx, posts); err != nil {
		store.Logger.Errorw("Could not create posts", "error", err.Error())
		return
	}

	store.Logger.Infof(" - Generating %d comments...\n", commentNum)
	comments := generateComments(commentNum, users, posts)
	if err := store.Comments.CreateBatch(ctx, comments); err != nil {
		store.Logger.Errorw("Could not create users", "error", err.Error())
		return
	}

	store.Logger.Infoln("Database seeding complete!")
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)

	wg := sync.WaitGroup{}
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()
			username := fmt.Sprintf("%s%d", usernames[i%len(usernames)], i)
			user := &store.User{
				Username: username,
				Email:    fmt.Sprintf("%s@example.com", username),
				IsActive: rand.Intn(2) == 0, // Randomly set IsActive to true or false
			}
			user.Password.Set(fmt.Sprintf("%sPassword", username))
			users[i] = user
		}()
	}
	wg.Wait()

	return users
}

func generateFollowers(num int, users []*store.User) []*store.Follower {
	followers := make([]*store.Follower, 0, num)
	existingPairs := []string{}

	activeUsers := make([]*store.User, 0)
	for _, user := range users {
		if user.IsActive {
			activeUsers = append(activeUsers, user)
		}
	}
	numActiveUsers := len(activeUsers)

	for len(followers) < num {
		user := activeUsers[rand.Intn(numActiveUsers)]
		follower := activeUsers[rand.Intn(numActiveUsers)]

		pairKey := fmt.Sprintf("%d-%d", user.ID, follower.ID)
		// Check if userID == followerID or the pair already exists
		if slices.Contains(existingPairs, pairKey) || user.ID == follower.ID {
			continue
		}

		// Add the new pair to existingPairs and followers
		existingPairs = append(existingPairs, pairKey)
		followers = append(followers, &store.Follower{
			UserID:     user.ID,
			FollowerID: follower.ID,
		})
	}

	return followers
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)

	wg := sync.WaitGroup{}
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()
			user := users[rand.Intn(len(users))]

			posts[i] = &store.Post{
				Title:   titles[rand.Intn(len(titles))],
				Content: contents[rand.Intn(len(contents))],
				Tags:    getRandomTags(),
				UserID:  user.ID,
			}
		}()
	}
	wg.Wait()
	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	cms := make([]*store.Comment, num)

	wg := sync.WaitGroup{}
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()
			cms[i] = &store.Comment{
				PostID:  posts[rand.Intn(len(posts))].ID,
				UserID:  users[rand.Intn(len(users))].ID,
				Content: comments[rand.Intn(len(comments))],
			}
		}()
	}
	wg.Wait()
	return cms
}

func getRandomTags() []string {
	numTags := rand.Intn(3) + 2 // Random number between 2 and 4

	selectedTags := make([]string, numTags)
	tagIndices := rand.Perm(len(tags))

	for i := 0; i < numTags; i++ {
		selectedTags[i] = tags[tagIndices[i]]
	}

	return selectedTags
}
